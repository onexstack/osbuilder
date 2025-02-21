package rid

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"os"
)

// Salt calculates the hash of the machine ID and returns it as a uint64 salt value.
func Salt() uint64 {
	// Use the FNV-1a hashing algorithm to calculate the hash value of the string.
	hasher := fnv.New64a()
	hasher.Write(ReadMachineID())

	// Convert the hash value into a uint64 salt.
	hashValue := hasher.Sum64()
	return hashValue
}

// ReadMachineID retrieves the machine ID. If it fails to obtain it, a random ID is generated.
func ReadMachineID() []byte {
	id := make([]byte, 3)
	machineID, err := readPlatformMachineID()
	if err != nil || len(machineID) == 0 {
		machineID, err = os.Hostname()
	}

	if err == nil && len(machineID) != 0 {
		hasher := sha256.New()
		hasher.Write([]byte(machineID))
		copy(id, hasher.Sum(nil))
	} else {
		// Fallback to generating a random number if the machine ID cannot be retrieved.
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("id: cannot get hostname nor generate a random number: %w; %w", err, randErr))
		}
	}
	return id
}

// readPlatformMachineID attempts to read the platform-specific machine ID.
func readPlatformMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil || len(data) == 0 {
		data, err = os.ReadFile("/sys/class/dmi/id/product_uuid")
	}
	return string(data), err
}
