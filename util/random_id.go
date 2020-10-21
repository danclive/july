package util

import (
	"github.com/danclive/nson-go"
)

// var pid = os.Getpid()
// var counter = uint32(time.Now().Nanosecond())
// var machineId = readMachineId()

// func readMachineId() []byte {
// 	id := make([]byte, 3)
// 	hostname, err1 := os.Hostname()
// 	if err1 != nil {
// 		_, err2 := io.ReadFull(rand.Reader, id)
// 		if err2 != nil {
// 			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
// 		}
// 		return id
// 	}
// 	hw := md5.New()
// 	hw.Write([]byte(hostname))
// 	copy(id, hw.Sum(nil))
// 	return id
// }

// func RandomID() string {
// 	var b [12]byte
// 	// Timestamp, 4 bytes, big endian
// 	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
// 	// Machine, first 3 bytes of md5(hostname)
// 	b[4] = machineId[0]
// 	b[5] = machineId[1]
// 	b[6] = machineId[2]
// 	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
// 	b[7] = byte(pid >> 8)
// 	b[8] = byte(pid)
// 	// Increment, 3 bytes, big endian
// 	i := atomic.AddUint32(&counter, 1)
// 	b[9] = byte(i >> 16)
// 	b[10] = byte(i >> 8)
// 	b[11] = byte(i)
// 	return fmt.Sprintf(`%x`, string(b[:]))
// }

func RandomID() string {
	return nson.NewMessageId().Hex()
}
