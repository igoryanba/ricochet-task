# Ricochet Installation

## Option A — Global install (symlink/copy)
```bash
# From repo root
./scripts/install-cli.sh            # installs to /usr/local/bin by default
# OR
./scripts/install-cli.sh "$HOME/.local/bin"
```

Then use globally:
```bash
ricochet-task --help
```

## Option B — npm-like local install (project script)
```bash
# Add project script (example)
echo '{"scripts":{"ricochet-task":"./ricochet-task"}}' > package.json
npm run ricochet-task -- --help
```

## Option C — Curl one-liner (symlink)
```bash
bash -c "cd /tmp && git clone https://github.com/YOUR_ORG/ricochet-task.git && cd ricochet-task && ./scripts/install-cli.sh"
```

## After install
```bash
# Quick setup
./scripts/install.sh && ./scripts/setup.sh
```

## Uninstall
```bash
rm -f /usr/local/bin/ricochet-task   # or the prefix you used
```
