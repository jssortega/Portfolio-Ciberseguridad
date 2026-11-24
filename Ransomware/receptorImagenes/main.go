package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func handleConn(c net.Conn) {
	defer c.Close()
	for {
		var size uint32
		if err := binary.Read(c, binary.BigEndian, &size); err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Error leyendo tamaño: %v", err)
			return
		}
		if size == 0 {
			continue
		}

		img := make([]byte, size)
		_, err := io.ReadFull(c, img)
		if err != nil {
			log.Printf("Error leyendo imagen (%d bytes): %v", size, err)
			return
		}
		
		filename := fmt.Sprintf("capture_%d.png", time.Now().UnixNano())
		if err := os.WriteFile(filename, img, 0644); err != nil {
			log.Printf("Error guardando %s: %v", filename, err)
			return
		}
	}
}

func main() {
	listenAddr := ":5555"
	// Aquí uso Listen y no Dial porque quiero escuchar conexiones no iniciarlas yo
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Error al escuchar %s: %v", listenAddr, err)
	}
	log.Printf("Escuchando en %s …", listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error al aceptar conexión: %v", err)
			continue
		}
		go handleConn(conn)
	}
}
