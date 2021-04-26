package order

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/decred/dcrd/crypto/blake256"
)

// Several types including OrderID and Commitment are defined as a Blake256
// hash. Define an internal hash type and hashSize for convenience.
const hashSize = blake256.Size // 32
type hash = [hashSize]byte

// OrderIDSize defines the length in bytes of an OrderID.
const OrderIDSize = hashSize

// OrderID is the unique identifier for each order. It is defined as the
// Blake256 hash of the serialized order.
type OrderID hash

// IDFromHex decodes an OrderID from a hexadecimal string.
func IDFromHex(sid string) (OrderID, error) {
	if len(sid) > OrderIDSize*2 {
		return OrderID{}, fmt.Errorf("invalid order ID. too long %d > %d", len(sid), OrderIDSize*2)
	}
	oidB, err := hex.DecodeString(sid)
	if err != nil {
		return OrderID{}, fmt.Errorf("order ID decode error: %w", err)
	}
	var oid OrderID
	copy(oid[OrderIDSize-len(oidB):], oidB)
	return oid, nil
}

// String returns a hexadecimal representation of the OrderID. String implements
// fmt.Stringer.
func (oid OrderID) String() string {
	return hex.EncodeToString(oid[:])
}

// MarshalJSON satisfies the json.Marshaller interface, and will marshal the
// id to a hex string.
func (oid OrderID) MarshalJSON() ([]byte, error) {
	return json.Marshal(oid.String())
}

// Bytes returns the order ID as a []byte.
func (oid OrderID) Bytes() []byte {
	return oid[:]
}

// Value implements the sql/driver.Valuer interface.
func (oid OrderID) Value() (driver.Value, error) {
	return oid[:], nil // []byte
}

// Scan implements the sql.Scanner interface.
func (oid *OrderID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		copy(oid[:], src)
		return nil
		//case string:
		// case nil:
		// 	*oid = nil
		// 	return nil
	}

	return fmt.Errorf("cannot convert %T to OrderID", src)
}

var zeroOrderID OrderID

// IsZero returns true if the order ID is zeros.
func (oid OrderID) IsZero() bool {
	return oid == zeroOrderID
}
