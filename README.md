<p align="center">
  <img width="250px" style="clip-path: circle()" src=".github/logo.webp" alt="libations logo">
</p>

<h1 align="center">Libations</h1>

This is a simple static website for hosting cocktail recipes. The actual site is rendered with Go templates
and served by default over Tailscale (using [tsnet]) by a simple Go binary that embeds the page.
The page is designed to be viewed on a mobile - it works _okay_ on bigger screens, but I've not yet
made that look "right".

The page is styled with the excellent [Vanilla Framework], because that's what I had to hand.

Cocktail recipes are served up from a JSON file containing the recipes. The format is listed in a
section below. There is an example [included](./static/sample.json).

<p align="center">
<img width="250px" src=".github/screenshot.png" alt="screenshot of libations"/>
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
  -recipes-file string
        path to a file containing drink recipes
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

Before building you must have [Go] installed.

```bash
git clone git@github.com:jnsgruk/libations

# Optional - if not provided you'll be prompted
export TS_AUTHKEY="tskey-auth-aBcdEfghIjKlMnOpQrStUvWxYz"

# Run the application on the tailnet
go run main.go

# Or run the application locally (handy for development)
go run main.go -local
```

## Recipe File Format

The [drinks.json](./static/sample.json) file is a list of JSON objects, where each object
represents a drink:

```json
[
  {
    "id": 5,
    "name": "Sazerum (Rum Sazerac)",
    "base": ["Rum"],
    "glass": ["9oz Lowball"],
    "method": ["Stir"],
    "ice": [],
    "ingredients": [
      { "name": "Rum", "measure": "50", "unit": "ml" },
      { "name": "Absinthe", "measure": "5", "unit": "ml" },
      { "name": "Dark Agave Syrup", "measure": "7.5", "unit": "ml" },
      { "name": "Peychaud Bitters", "measure": "2.5", "unit": "ml" },
      { "name": "Lemon Juice", "measure": "2", "unit": "dash" }
    ],
    "garnish": ["Lemon Twist"],
    "notes": "Preferred rum is Plantation XO. Chill glass before and rinse with 5ml Absinthe"
  }
]
```

[Go]: https://go.dev/
[tsnet]: https://tailscale.com/kb/1244/tsnet/
[Vanilla Framework]: https://vanillaframework.io
[Canonical]: https://canonical.com
