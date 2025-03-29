package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

type captureWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *captureWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func GzipWithMetrics(paths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		match := false
		for _, p := range paths {
			if strings.HasPrefix(c.FullPath(), p) {
				match = true
				break
			}
		}

		if !match {
			c.Next()
			return
		}

		if c.GetHeader("Accept-Encoding") != "gzip" {
			fmt.Printf("[Sin compresión] %s %s - Cliente no acepta gzip\n", c.Request.Method, c.FullPath())
			c.Next()
			return
		}

		buffer := new(bytes.Buffer)
		originalWriter := &captureWriter{ResponseWriter: c.Writer, buf: buffer}
		c.Writer = originalWriter

		c.Next()

		// Comprimir la respuesta capturada
		var compressed bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressed)
		_, _ = gzipWriter.Write(buffer.Bytes())
		gzipWriter.Close()

		// Calcular tamaños
		originalSize := buffer.Len()
		compressedSize := compressed.Len()

		// Agregar headers
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", compressedSize))

		// Calcular y registrar ratio
		if originalSize > 0 {
			ratio := float64(compressedSize) / float64(originalSize)
			ratioStr := fmt.Sprintf("%.2f", ratio)
			c.Writer.Header().Set("X-Compression-Ratio", ratioStr)

			// Loguear en consola
			fmt.Printf("[Compresión] %s %s - Original: %d bytes | Comprimido: %d bytes | Ratio: %s\n",
				c.Request.Method, c.Request.URL.Path, originalSize, compressedSize, ratioStr)
		}

		// Escribir la respuesta comprimida en el ResponseWriter original
		c.Writer.WriteHeaderNow()
		_, _ = originalWriter.ResponseWriter.Write(compressed.Bytes())
	}
}
