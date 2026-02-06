#version 300 es
precision highp float;

// Text texture (braille rendered to 2D canvas)
uniform sampler2D uText;
// Time in frames (monotonic counter)
uniform float uTime;
// Muted color (base)
uniform vec3 uMuted;
// Foreground color (peak)
uniform vec3 uFg;
// Grid dimensions (cols, rows)
uniform vec2 uGrid;

in vec2 vUV;
out vec4 fragColor;

// --- Layer constants (exact port from shimmer.go) ---

// Layer 1 (primary): medium-scale blobs
const float L1_X_SCALE = 0.14;
const float L1_Y_SCALE = 0.22;
const float L1_TIME_SCALE = 0.004;
const float L1_THRESH_CENTER = 0.52;
const float L1_THRESH_RADIUS = 0.18;
const float L1_WEIGHT = 1.0;

// Layer 2 (wash): large slow ambient glow
const float L2_X_SCALE = 0.07;
const float L2_Y_SCALE = 0.1;
const float L2_TIME_SCALE = 0.002;
const float L2_TIME_OFFSET = 80.0;
const float L2_THRESH_CENTER = 0.5;
const float L2_THRESH_RADIUS = 0.25;
const float L2_WEIGHT = 0.35;

// Layer 3 (detail): small bright speckles
const float L3_X_SCALE = 0.25;
const float L3_Y_SCALE = 0.35;
const float L3_TIME_SCALE = 0.006;
const float L3_TIME_OFFSET = 160.0;
const float L3_THRESH_CENTER = 0.58;
const float L3_THRESH_RADIUS = 0.12;
const float L3_WEIGHT = 0.3;

// Breathing
const float BREATH_BASE = 0.7;
const float BREATH_AMP = 0.3;
const float BREATH_FREQ = 0.010;

// --- Noise functions ---

// Deterministic hash for integer lattice point → [-0.5, 0.5)
// Port of Go: h = uint32(x*374761393+y*668265263+z*1440670441) ^ 0x27d4eb2d
float latticeHash(int x, int y, int z) {
    // Use uint arithmetic to match Go's uint32 behavior
    uint h = uint(x) * 374761393u + uint(y) * 668265263u + uint(z) * 1440670441u;
    h ^= 0x27d4eb2du;
    h = (h ^ (h >> 13u)) * 1274126177u;
    h = h ^ (h >> 16u);
    return float(h & 0x7fffffffu) / float(0x80000000u) - 0.5;
}

// Trilinear interpolated value noise in ~[-0.5, 0.5]
float smoothNoise3D(vec3 p) {
    ivec3 i = ivec3(floor(p));
    vec3 f = fract(p);
    // Smoothstep interpolation
    f = f * f * (3.0 - 2.0 * f);

    float c000 = latticeHash(i.x,     i.y,     i.z);
    float c100 = latticeHash(i.x + 1, i.y,     i.z);
    float c010 = latticeHash(i.x,     i.y + 1, i.z);
    float c110 = latticeHash(i.x + 1, i.y + 1, i.z);
    float c001 = latticeHash(i.x,     i.y,     i.z + 1);
    float c101 = latticeHash(i.x + 1, i.y,     i.z + 1);
    float c011 = latticeHash(i.x,     i.y + 1, i.z + 1);
    float c111 = latticeHash(i.x + 1, i.y + 1, i.z + 1);

    float x0 = mix(c000, c100, f.x);
    float x1 = mix(c010, c110, f.x);
    float x2 = mix(c001, c101, f.x);
    float x3 = mix(c011, c111, f.x);

    float y0 = mix(x0, x1, f.y);
    float y1 = mix(x2, x3, f.y);

    return mix(y0, y1, f.z);
}

// Fractal Brownian motion: 3 octaves → [0, 1]
float fbmNoise(vec3 p) {
    float v = 0.0;
    float amp = 0.5;
    float freq = 1.0;
    for (int i = 0; i < 3; i++) {
        v += amp * smoothNoise3D(p * freq);
        freq *= 2.0;
        amp *= 0.5;
    }
    return v + 0.5;
}

// Soft threshold → [0, 1]
float smoothThreshold(float value, float center, float radius) {
    float low = center - radius;
    float high = center + radius;
    return smoothstep(low, high, value);
}

// --- Brightness computation (exact port of shimmer.go brightnessAt) ---

float brightnessAt(float row, float col, float t) {
    float r = row;
    float c = col;

    // Layer 1 per-row drift: 4 incommensurate sinusoids
    float drift = 3.0 * sin(t * 0.006 + r * 0.41)
                + 2.0 * sin(t * 0.011 + r * 0.67)
                + 1.5 * sin(t * 0.003 + r * 0.23)
                + 1.0 * sin(t * 0.017 + r * 1.1);

    float nx = (c + drift) * L1_X_SCALE;
    float ny = r * L1_Y_SCALE;
    float nz = t * L1_TIME_SCALE;
    float n1 = fbmNoise(vec3(nx, ny, nz));
    float b1 = smoothThreshold(n1, L1_THRESH_CENTER, L1_THRESH_RADIUS);

    // Layer 2: large slow wash
    float drift2 = 2.0 * sin(t * 0.004 + r * 0.3)
                 + 1.5 * sin(t * 0.009 + r * 0.55);
    float nx2 = (c + drift2) * L2_X_SCALE;
    float ny2 = r * L2_Y_SCALE;
    float n2 = fbmNoise(vec3(nx2, ny2, t * L2_TIME_SCALE + L2_TIME_OFFSET));
    float b2 = smoothThreshold(n2, L2_THRESH_CENTER, L2_THRESH_RADIUS) * L2_WEIGHT;

    // Layer 3: fine detail speckles
    float drift3 = 2.5 * sin(t * 0.014 + r * 0.8)
                 + 1.0 * sin(t * 0.008 + r * 0.35);
    float nx3 = (c + drift3) * L3_X_SCALE;
    float ny3 = r * L3_Y_SCALE;
    float n3 = fbmNoise(vec3(nx3, ny3, t * L3_TIME_SCALE + L3_TIME_OFFSET));
    float b3 = smoothThreshold(n3, L3_THRESH_CENTER, L3_THRESH_RADIUS) * L3_WEIGHT;

    float combined = b1 + b2 + b3;

    // Global breathing
    float breath = BREATH_BASE + BREATH_AMP * sin(t * BREATH_FREQ);
    combined *= breath;

    return min(combined, 1.0);
}

void main() {
    // Sample text texture alpha
    float alpha = texture(uText, vUV).a;

    // Discard transparent pixels (empty braille / background)
    if (alpha < 0.01) discard;

    // Quantize UV to cell grid coordinates
    float col = vUV.x * uGrid.x;
    float row = vUV.y * uGrid.y;

    float brightness = brightnessAt(row, col, uTime);

    vec3 color = mix(uMuted, uFg, brightness);
    fragColor = vec4(color, alpha);
}
