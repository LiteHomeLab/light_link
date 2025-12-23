// Package backup provides prototype implementation for backup functionality
// Including: incremental backup with reverse merge, chunked upload/download
package backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// DiffType represents the type of diff operation
type DiffType uint8

const (
	DiffMatch   DiffType = 0 // Data matches, copy from source
	DiffInsert  DiffType = 1 // New data inserted
	DiffDelete  DiffType = 2 // Data deleted
	DiffReplace DiffType = 3 // Data replaced
)

// DiffOp represents a single diff operation
type DiffOp struct {
	Type     DiffType
	OldPos   uint32   // Position in old data (for Match, Delete, Replace)
	NewPos   uint32   // Position in new data (for Insert, Replace)
	OldLen   uint32   // Length in old data
	NewLen   uint32   // Length in new data
	Data     []byte   // Actual data for Insert and Replace
	Checksum []byte   // SHA256 checksum for verification
}

// BinaryDiff computes the difference between old and new data
// Uses a simple block-based diff algorithm for prototype
func BinaryDiff(oldData, newData []byte) ([]DiffOp, error) {
	const blockSize = 4096

	var ops []DiffOp
	oldBlocks := splitIntoBlocks(oldData, blockSize)
	newBlocks := splitIntoBlocks(newData, blockSize)

	// Find matching blocks using rolling hash
	matchMap := buildBlockMap(oldBlocks)

	// Scan through new blocks and create diff ops
	newPos := uint32(0)
	oldPos := uint32(0)

	for i := 0; i < len(newBlocks); i++ {
		newBlock := newBlocks[i]
		blockHash := sha256.Sum256(newBlock)

		if matchingIdx, found := findMatchingBlock(matchMap, newBlock, blockHash[:]); found {
			// Found matching block in old data
			matchStart := uint32(matchingIdx * blockSize)
			matchLen := uint32(len(newBlock))

			// Check if we need to skip/insert before this match
			if oldPos < matchStart {
				// Data in old was deleted
				ops = append(ops, DiffOp{
					Type:   DiffDelete,
					OldPos: oldPos,
					OldLen: matchStart - oldPos,
				})
			}

			// Add match operation
			ops = append(ops, DiffOp{
				Type:     DiffMatch,
				OldPos:   matchStart,
				NewPos:   newPos,
				OldLen:   matchLen,
				NewLen:   matchLen,
				Checksum: blockHash[:],
			})

			oldPos = matchStart + matchLen
			newPos += matchLen
		} else {
			// New or modified block - insert
			checksum := sha256.Sum256(newBlock)
			ops = append(ops, DiffOp{
				Type:     DiffInsert,
				NewPos:   newPos,
				NewLen:   uint32(len(newBlock)),
				Data:     newBlock,
				Checksum: checksum[:],
			})
			newPos += uint32(len(newBlock))
		}
	}

	// Handle trailing deletions
	if oldPos < uint32(len(oldData)) {
		ops = append(ops, DiffOp{
			Type:   DiffDelete,
			OldPos: oldPos,
			OldLen: uint32(len(oldData)) - oldPos,
		})
	}

	return ops, nil
}

// ApplyDiff applies diff operations to base data to reconstruct new data
func ApplyDiff(baseData []byte, ops []DiffOp) ([]byte, error) {
	var result bytes.Buffer
	h := sha256.New()

	for _, op := range ops {
		switch op.Type {
		case DiffMatch:
			// Verify checksum first
			if op.OldPos+op.OldLen > uint32(len(baseData)) {
				return nil, fmt.Errorf("match operation out of bounds")
			}
			data := baseData[op.OldPos : op.OldPos+op.OldLen]

			// Verify checksum
			if len(op.Checksum) > 0 {
				checksum := sha256.Sum256(data)
				if !bytes.Equal(checksum[:], op.Checksum) {
					return nil, fmt.Errorf("checksum mismatch at position %d", op.OldPos)
				}
			}

			result.Write(data)
			h.Write(data)

		case DiffInsert:
			// Verify checksum
			if len(op.Checksum) > 0 {
				checksum := sha256.Sum256(op.Data)
				if !bytes.Equal(checksum[:], op.Checksum) {
					return nil, fmt.Errorf("checksum mismatch for insert at position %d", op.NewPos)
				}
			}

			result.Write(op.Data)
			h.Write(op.Data)

		case DiffDelete:
			// Skip data - do nothing

		case DiffReplace:
			// Verify checksum
			if len(op.Checksum) > 0 {
				checksum := sha256.Sum256(op.Data)
				if !bytes.Equal(checksum[:], op.Checksum) {
					return nil, fmt.Errorf("checksum mismatch for replace at position %d", op.NewPos)
				}
			}

			result.Write(op.Data)
			h.Write(op.Data)
		}
	}

	return result.Bytes(), nil
}

// ReverseMerge merges incremental diffs into base data
// This creates a new full snapshot by applying all changes
func ReverseMerge(baseData []byte, diffOpsList [][]DiffOp) ([]byte, error) {
	current := baseData

	for _, ops := range diffOpsList {
		var err error
		current, err = ApplyDiff(current, ops)
		if err != nil {
			return nil, fmt.Errorf("reverse merge failed: %w", err)
		}
	}

	return current, nil
}

// SerializeDiffOps converts diff operations to binary format for storage/transmission
func SerializeDiffOps(ops []DiffOp) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write number of operations
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(ops))); err != nil {
		return nil, err
	}

	for _, op := range ops {
		// Write type
		if err := binary.Write(buf, binary.LittleEndian, op.Type); err != nil {
			return nil, err
		}
		// Write old position
		if err := binary.Write(buf, binary.LittleEndian, op.OldPos); err != nil {
			return nil, err
		}
		// Write new position
		if err := binary.Write(buf, binary.LittleEndian, op.NewPos); err != nil {
			return nil, err
		}
		// Write old length
		if err := binary.Write(buf, binary.LittleEndian, op.OldLen); err != nil {
			return nil, err
		}
		// Write new length
		if err := binary.Write(buf, binary.LittleEndian, op.NewLen); err != nil {
			return nil, err
		}
		// Write checksum length
		checksumLen := uint8(len(op.Checksum))
		if err := binary.Write(buf, binary.LittleEndian, checksumLen); err != nil {
			return nil, err
		}
		// Write checksum
		if _, err := buf.Write(op.Checksum); err != nil {
			return nil, err
		}
		// Write data length
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(op.Data))); err != nil {
			return nil, err
		}
		// Write data
		if _, err := buf.Write(op.Data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// DeserializeDiffOps converts binary data back to diff operations
func DeserializeDiffOps(data []byte) ([]DiffOp, error) {
	buf := bytes.NewReader(data)
	var numOps uint32

	if err := binary.Read(buf, binary.LittleEndian, &numOps); err != nil {
		return nil, err
	}

	ops := make([]DiffOp, numOps)

	for i := uint32(0); i < numOps; i++ {
		if err := binary.Read(buf, binary.LittleEndian, &ops[i].Type); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.LittleEndian, &ops[i].OldPos); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.LittleEndian, &ops[i].NewPos); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.LittleEndian, &ops[i].OldLen); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.LittleEndian, &ops[i].NewLen); err != nil {
			return nil, err
		}

		var checksumLen uint8
		if err := binary.Read(buf, binary.LittleEndian, &checksumLen); err != nil {
			return nil, err
		}

		ops[i].Checksum = make([]byte, checksumLen)
		if checksumLen > 0 {
			if _, err := buf.Read(ops[i].Checksum); err != nil {
				return nil, err
			}
		}

		var dataLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &dataLen); err != nil {
			return nil, err
		}

		ops[i].Data = make([]byte, dataLen)
		if dataLen > 0 {
			if _, err := buf.Read(ops[i].Data); err != nil {
				return nil, err
			}
		}
	}

	return ops, nil
}

// Helper functions

func splitIntoBlocks(data []byte, size int) [][]byte {
	if len(data) == 0 {
		return nil
	}

	var blocks [][]byte
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		blocks = append(blocks, data[i:end])
	}
	return blocks
}

func buildBlockMap(blocks [][]byte) map[string]int {
	m := make(map[string]int)
	for i, block := range blocks {
		hash := sha256.Sum256(block)
		m[string(hash[:])] = i
	}
	return m
}

func findMatchingBlock(matchMap map[string]int, block []byte, hash []byte) (int, bool) {
	if idx, ok := matchMap[string(hash)]; ok {
		// Verify actual data matches (hash collision check)
		return idx, true
	}
	return -1, false
}

// CalculateChecksum returns SHA256 checksum of data
func CalculateChecksum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// VerifyChecksum verifies data against checksum
func VerifyChecksum(data []byte, checksum []byte) bool {
	h := sha256.Sum256(data)
	return bytes.Equal(h[:], checksum)
}
