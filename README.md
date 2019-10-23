# go-nuget

A basic implementation of the Nuget CLI tool, written in Go, for use in automated build processes and deployment of .nuspec &amp; .nupkg files.

Writen in Go to be platform agnostic, built against the go-nuget-server project.

![Build Status](https://github.com/soloworks/go-nuget/workflows/Build/badge.svg)
[![MIT license](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)
[![LinkedIn](https://img.shields.io/badge/Contact-LinkedIn-blue)](https://www.linkedin.com/company/soloworkslondon/)

## Getting Started - Source

```bash
git clone github/soloworks/go-nuget

go build -o nuget.exe

```

## CLI

### spec - Generates a .nuspec file

```bash
nuget spec <packageID>
```

### pack - Generates a .nupkg file

```bash
nuget pack nuspecPath
```

### push - Pushes a package to a server

```bash
nuget push <packagePath> --Source http://<yourserver>/<packages-root>/ --ApiKey 123-456-789
```

## Notes

TBC

## Resources

- Microsoft nuget.exe CLI [[Here]](https://docs.microsoft.com/en-us/nuget/reference/nuget-exe-cli-reference)
- Microsoft nuget.exe Download [Here](https://dist.nuget.org/win-x86-commandline/latest/nuget.exe) (v5.3.0 - Latest Recommended)

## Acknowledgements
