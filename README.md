# Libations

This is a simple static website for hosting cocktail recipes. The actual site is built using [Hugo]
and served by default over Tailscale (using [tsnet]) by a simple Go binary that embeds the page.
The page is designed to be viewed on a mobile - it works _okay_ on bigger screens, but I've not yet
made that look "right".

The page is styled with the excellent [Vanilla Framework], because that's what I had to hand.

Cocktail recipes are served up from a JSON file containing the recipes. The format is listed in a
section below. There is an example [included](./webui/data/drinks.json).

<p align="center">
<img src=".github/screenshot.png" alt="screenshot of libations" style="max-height:500px"/>
</p>

## Usage

### Command line flags

```
./libations -help
Usage of /bin/libations:
  -addr string
        the address to listen on in the case of a local listener (default ":8080")
  -hostname string
        hostname to use on the tailnet (default "libations")
  -local
        start on local addr; don't attach to a tailnet
  -tsnet-logs
        include tsnet logs in application logs (default true)
```

### Using Nix

```bash
# Enter the development shell
nix develop

# Run the package
nix run .#libations
```

### Otherwise...

Before building you must have [Hugo], and [Go] installed.

```bash
git clone git@github.com:jnsgruk/libations

# Build the static site with Hugo
go generate

# Optional - if not provided you'll be prompted
export TS_AUTHKEY="tskey-auth-aBcdEfghIjKlMnOpQrStUvWxYz"

# Run the application
go run main.go
```

For iterating on the web interface, it can be quicker to just use Hugo:

```bash
cd webui

hugo serve --minify
```

## Recipe File Format

The [drinks.json](./webui/data/drinks.json) file is a list of JSON objects, where each object
represents a drink:

```json
[
  {
    "id": 7,
    "name": "Adderly",
    "base": ["Bourbon"],
    "glass": ["Coup"],
    "method": ["Shake"],
    "ice": [],
    "ingredients": [
      { "name": "Orange Bitters", "measure": "1", "unit": "dash" },
      { "name": "Lemon", "measure": "20", "unit": "ml" },
      { "name": "Sugar", "measure": "5", "unit": "ml" },
      { "name": "Maraschino", "measure": "15", "unit": "ml" },
      { "name": "Bourbon", "measure": "45", "unit": "ml" }
    ],
    "garnish": [],
    "notes": "I like this one, good with frozen Bourbon"
  }
]
```

[Hugo]: https://gohugo.io
[Go]: https://go.dev/
[tsnet]: https://tailscale.com/kb/1244/tsnet/
[Vanilla Framework]: https://vanillaframework.io
[Canonical]: https://canonical.com
