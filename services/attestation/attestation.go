// +build sgx_enclave

package attestation

/*
#cgo LDFLAGS: -L../../sgx/target/ -ladapters
#include <stdlib.h>
#include "../../sgx/libadapters/adapters.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Quote passes an arbitrary string into the enclave and returns a remote
// attestation
func Quote(input string) (string, error) {
	cInput := C.CString(string(input))
	defer C.free(unsafe.Pointer(cInput))

	buffer := make([]byte, 8192)
	output := (*C.char)(unsafe.Pointer(&buffer[0]))
	bufferCapacity := C.int(len(buffer))
	outputLen := C.int(0)
	outputLenPtr := (*C.int)(unsafe.Pointer(&outputLen))

	if _, err := C.quote(cInput, output, bufferCapacity, outputLenPtr); err != nil {
		return "", fmt.Errorf("SGX quote: %v", err)
	}

	return C.GoStringN(output, outputLen), nil
}
