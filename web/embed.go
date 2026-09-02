package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// DistFS devuelve un sistema de archivos para servir estáticos que apunta directamente al interior de "dist".
func DistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
