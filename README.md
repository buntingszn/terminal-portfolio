# terminal-portfolio

A portfolio you can browse in a terminal or a browser. Both versions render from the same shared JSON data.

**Web** — [kpm.fyi](https://kpm.fyi)
**Terminal** — [ssh.kpm.fyi](https://ssh.kpm.fyi)

## Architecture

```
terminal-portfolio/
├── data/         Shared JSON content (consumed by both web and tui)
├── web/          Astro static site
├── tui/          Go + Bubbletea TUI served over SSH via Wish
└── scripts/      Validation utilities
```

The web version is a static Astro site. The terminal version is a Bubbletea TUI served over SSH with [Wish](https://github.com/charmbracelet/wish), made browser-accessible through [ttyd](https://github.com/nicm/ttyd) and a Cloudflare Tunnel.

## Stack

| Layer | Technology                          |
| ----- | ----------------------------------- |
| Web   | Astro, TypeScript, CSS              |
| TUI   | Go, Bubbletea, Lipgloss, Wish       |
| Data  | Shared JSON with validation scripts |
| Infra | Cloudflare Tunnel, ttyd, systemd    |

## License

MIT
