# CLAUDE.md

Guía para Claude Code (claude.ai/code) al trabajar en este repositorio.

## Qué es mnemonic

Gestión de conocimiento organizacional con búsqueda semántica, en Go sobre ChromaDB.
Se distribuye por **dos canales a la vez**, y eso es lo que más fácil se pasa por alto:

1. El **binario** `mnemonic`, que la gente instala con `scripts/install.sh`.
2. El **plugin de Claude Code**, que se instala con
   `claude plugin marketplace add marioser/mnemonic`.

Una release toca los dos. Publicar uno sin el otro deja la instalación a medias.

## Regla de release (la más importante de este archivo)

**Todo cambio que entra a `main` se publica con un tag, y la release tiene que traer los
cuatro artefactos: Linux y macOS, `amd64` y `arm64`.**

El motivo no es estético. `scripts/install.sh` detecta OS/arch y baja el binario de la
**última release de GitHub**. Si no hay release, lo que está mergeado en `main` **no le
llega a nadie**: quien instale se sigue llevando la versión anterior, con el bug que ya
arreglaste.

Esto ya pasó. El 2026-09-29 se mergeó el fix del PK-ID (`#2`) y el único tag siguió
siendo `v0.1.0` durante días. Todas las instalaciones nuevas en ese lapso se llevaron el
binario sin el arreglo.

### Dónde NO está el problema

La matriz de compilación **ya está completa** en `.goreleaser.yaml`:

```yaml
goos:   [linux, darwin]
goarch: [amd64, arm64]
```

Son las cuatro combinaciones. Nunca hizo falta agregar Linux: sale solo.

El cuello es el tag. `.github/workflows/release.yaml` arranca con:

```yaml
on:
  push:
    tags: ['v*']
```

**Sin tag no corre goreleaser, y sin goreleaser no hay binarios nuevos.** Es el único
paso manual de toda la cadena, y es justamente el que se olvida.

### El ciclo, con los comandos

```bash
# 1. Mergear a main como siempre (issue → worktree → PR → merge).

# 2. Etiquetar. SemVer; el tag es lo que dispara todo.
git checkout main && git pull
git tag v0.2.0
git push origin v0.2.0

# 3. Verificar que la release trajo los 4 artefactos.
gh release view v0.2.0 --json assets --jq '.assets[].name'
```

El paso 3 no es ceremonia: si goreleaser falla a mitad, el tag queda igual y la release
sale incompleta o vacía. Mirar los assets es la única confirmación real.

### Verificar la matriz sin publicar

Antes de etiquetar, para no descubrir un error de compilación cruzada con el tag ya
puesto:

```bash
goreleaser build --snapshot --clean   # compila las 4 combinaciones, no publica
```

**`goreleaser` no viene instalado**: `brew install goreleaser`. Si no lo querés instalar,
la compilación cruzada suelta alcanza para detectar lo grueso:

```bash
GOOS=linux  GOARCH=amd64 go build ./cmd/mnemonic
GOOS=darwin GOARCH=arm64 go build ./cmd/mnemonic
```

Cruza sin toolchain de C porque el árbol de dependencias es Go puro —`pure-onnx`,
`pure-tokenizers` y `purego` están elegidos precisamente para eso—. Si algún día entra
una dependencia con CGO, la compilación cruzada se rompe y hay que resolverlo **antes**
de etiquetar, no después.

## Comandos

```bash
go build ./...                 # compilar todo
go test ./... -short           # lo que corre CI; -short saltea lo que necesita ChromaDB
go test ./...                  # suite completa: requiere un ChromaDB accesible
go vet ./...                   # CI también lo corre, y falla el build
go build -o mnemonic ./cmd/mnemonic
```

CI (`.github/workflows/ci.yaml`) corre en cada push a `main` y en cada PR: build, test
`-short` y vet. Los tres tienen que pasar.

## Estructura

| Carpeta | Qué hay |
|---|---|
| `cmd/mnemonic/` | Entrada del binario; los subcomandos cuelgan de `commands/` (cobra) |
| `internal/chroma/` | Cliente de ChromaDB |
| `internal/embeddings/` | Generación de embeddings |
| `internal/mcp/` | Servidor MCP (`mcp-go`) |
| `internal/http/` | API HTTP (`chi`) |
| `internal/domains/` | Dominios de conocimiento |
| `internal/config/` | Config; el default vive en `config/default.yaml` |
| `internal/sync/` | Sincronización |
| `plugin/` | El plugin de Claude Code que viaja dentro del archivo de la release |

`.goreleaser.yaml` empaqueta `plugin/**/*` y `config/default.yaml` **dentro** del
`tar.gz`. Si agregás un archivo que el plugin necesita en runtime y no está bajo esas
rutas, compila, la release sale, y el plugin falla en la máquina del usuario.

## Versión

`.goreleaser.yaml` inyecta la versión por `ldflags` contra
`cmd/mnemonic/commands`:

```
-X .../commands.Version={{.Version}}
-X .../commands.Commit={{.ShortCommit}}
-X .../commands.BuildDate={{.Date}}
```

Los valores por defecto están en `cmd/mnemonic/commands/root.go`:

```go
Version   = "dev"
Commit    = "none"
BuildDate = "unknown"
```

Un binario compilado a mano con `go build` se queda con esos. Es esperable en desarrollo;
si un usuario reporta `dev`, significa que **no instaló desde la release** y cualquier
diagnóstico de versión sobre su máquina no sirve.

## Convenciones

- **Commits**: conventional commits, en español (`fix(references): …`). Sin atribución de
  IA, nunca.
- **Flujo**: issue → worktree → PR → merge → **tag**. El tag es parte del flujo, no un
  extra opcional.
- **Go**: el módulo fija `go 1.24.11`; CI usa `go-version: '1.24'`. Si subís el mínimo en
  `go.mod`, subí también las dos workflows.
