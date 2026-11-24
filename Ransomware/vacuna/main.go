package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------
// 1. Buscar los archivos igual que el ransom
func buscarArchivos(raiz string) ([]string, error) {
	var archivos []string
	extensiones := map[string]bool{
		".jpg": true, ".docx": true, ".xlsx": true,
		".pptx": true, ".pdf": true, ".png": true,
	}
	raiz, err := filepath.Abs(raiz)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if extensiones[strings.ToLower(filepath.Ext(d.Name()))] {
			archivos = append(archivos, ruta)
		}
		return nil
	})
	return archivos, err
}

// ---------------------------------------------------------------------
// 2. Función de descifrado (AES‑GCM), hace lo contrario
func descifrarAES(cifrado []byte, clave []byte) ([]byte, error) {
	block, err := aes.NewCipher(clave)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()

	nonce, data := cifrado[:nonceSize], cifrado[nonceSize:]

	return gcm.Open(nil, nonce, data, nil)
}

// ---------------------------------------------------------------------
func main() {
	// 1. Buscar archivos
	archivos, err := buscarArchivos(`C:\Users`)
	if err != nil {
		fmt.Println("Error al buscar archivos:", err)
		return
	}
	fmt.Printf("Se encontraron %d archivos a cifrar.\n", len(archivos))

	// 2. La clave ahora se la tengo que pasar por terminal

	lector := bufio.NewReader(os.Stdin)

	fmt.Print("Introduce la clave de descifrado: ")

	clave, err := lector.ReadString('\n')
	if err != nil {
		fmt.Println("Error leyendo la entrada:", err)
		return
	}

	decoded, errHex := hex.DecodeString(clave)
	if errHex == nil {
		return
	}

	// 3. Descifra cada archivo
	for _, ruta := range archivos {
		plain, err := os.ReadFile(ruta)
		if err != nil {
			fmt.Println("No se pudo leer", ruta, ":", err)
			continue
		}
		cifrado, err := descifrarAES(plain, decoded)
		if err != nil {
			fmt.Println("Error descifrando", ruta, ":", err)
			continue
		}
		// sobrescribir
		if err = os.WriteFile(ruta, cifrado, 0600); err != nil {
			fmt.Println("No se pudo escribir", ruta, ":", err)
			continue
		}
	}

	fmt.Println("Proceso concluido.")
}
