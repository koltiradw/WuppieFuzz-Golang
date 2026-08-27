package wgca

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"unsafe"
)

/*
#include <stdint.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

static struct {
	uint8_t *start;
	size_t size;
} CovMap = {0};


__attribute__((weak))
void __sanitizer_cov_trace_cmp1(uint8_t Arg1, uint8_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_8bit_counters_init(uint8_t *Start, uint8_t *Stop) {
        if (Start == NULL || Stop == NULL || Stop <= Start) {
        	return;
    	}
	CovMap.start = Start;
	CovMap.size  = (size_t)(Stop - Start);
}

__attribute__((weak))
uint8_t* get_coverage_map_start(void) {
	return CovMap.start;
}

__attribute__((weak))
size_t get_coverage_map_size(void) {
	return CovMap.size;
}

__attribute__((weak))
void copy_and_reset(uint8_t *dst, size_t dst_size, int do_reset) {
	if (CovMap.start == NULL) {
		return;
	}
	if (dst != NULL && dst_size > 0) {
		size_t n = CovMap.size;
		if (n > dst_size) {
			n = dst_size;
		}
		memcpy(dst, CovMap.start, n);
	}
	if (do_reset) {
		memset(CovMap.start, 0, CovMap.size);
	}
}
__attribute__((weak))
void __sanitizer_cov_trace_cmp2(uint16_t Arg1, uint16_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_cmp4(uint32_t Arg1, uint32_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_cmp8(uint64_t Arg1, uint64_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_const_cmp1(uint8_t Arg1, uint8_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_const_cmp2(uint16_t Arg1, uint16_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_const_cmp4(uint32_t Arg1, uint32_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_trace_const_cmp8(uint64_t Arg1, uint64_t Arg2) {
}
__attribute__((weak))
void __sanitizer_cov_pcs_init(const uintptr_t *pcs_beg,
                              const uintptr_t *pcs_end) {
}
__attribute__((weak))
void __sanitizer_weak_hook_strcmp(void *caller_pc, const char *s1,
                                   const char *s2, int result) {
}
*/
import "C"

const HOST = "0.0.0.0"
const TYPE = "tcp"
const DEFAULT_PORT = "1337"

const HEADER_SIZE = 10

var MAGIC = []byte("WGCA") // 0x57 0x47 0x43 0x41

const VERSION byte = 0x01

const MSG_REQUEST_DUMP byte = 0x01
const MSG_RESPONSE_DUMP byte = 0x02

func handleRequest(conn net.Conn) error {
	header := make([]byte, HEADER_SIZE)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	if !bytes.Equal(header[0:4], MAGIC) {
		return fmt.Errorf("invalid magic: % x", header[0:4])
	}

	msgType := header[5]
	length := binary.LittleEndian.Uint32(header[6:10])

	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(conn, payload); err != nil {
			return err
		}
	}

	switch msgType {
	case MSG_REQUEST_DUMP:
		return handleDump(conn, payload)
	default:
		return fmt.Errorf("unknown message type: 0x%02x", msgType)
	}
}

func handleDump(conn net.Conn, payload []byte) error {
	var resetByte byte
	if len(payload) > 0 {
		resetByte = payload[0]
	}

	covMap := make([]byte, int(C.get_coverage_map_size()))
	var dst *C.uint8_t
	if len(covMap) > 0 {
		dst = (*C.uint8_t)(unsafe.Pointer(&covMap[0]))
	}
	C.copy_and_reset(dst, C.size_t(len(covMap)), C.int(resetByte))

	return writeFrame(conn, MSG_RESPONSE_DUMP, covMap)
}

func writeFrame(conn net.Conn, msgType byte, payload []byte) error {
	frame := make([]byte, HEADER_SIZE+len(payload))
	copy(frame[0:4], MAGIC)
	frame[4] = VERSION
	frame[5] = msgType
	binary.LittleEndian.PutUint32(frame[6:10], uint32(len(payload)))
	copy(frame[HEADER_SIZE:], payload)
	_, err := conn.Write(frame)
	return err
}

func init() {
	go startCoverageServer()
}

func resolvePort() string {
	if port := os.Getenv("WUPPIE_COVERAGE_PORT"); port != "" {
		return port
	}
	return DEFAULT_PORT
}

func startCoverageServer() {
	listen, err := net.Listen(TYPE, net.JoinHostPort(HOST, resolvePort()))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Printf("coverage agent listening on %s", listen.Addr())

	defer listen.Close()

	conn, err := listen.Accept()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for {
		if err := handleRequest(conn); err != nil {
			conn.Close()
			conn, err = listen.Accept()
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}
	}
}
