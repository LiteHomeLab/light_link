package backup

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ChunkSize is the default size for data chunks (1MB)
const ChunkSize = 1024 * 1024

// Chunk represents a data chunk with metadata
type Chunk struct {
	Index    uint32 // Chunk index (0-based)
	Data     []byte // Actual chunk data
	Checksum []byte // SHA256 checksum
	Size     uint32 // Actual data size (may be less than ChunkSize for last chunk)
}

// ChunkMetadata contains metadata about chunked data transfer
type ChunkMetadata struct {
	TotalChunks uint32 // Total number of chunks
	TotalSize   uint64 // Total data size
	FileID      string // Unique file identifier
	Checksum    []byte // Checksum of entire data
}

// ChunkSplitter splits data into chunks for upload
type ChunkSplitter struct {
	data      []byte
	chunkSize int
	metadata  ChunkMetadata
}

// NewChunkSplitter creates a new chunk splitter
func NewChunkSplitter(data []byte) *ChunkSplitter {
	return &ChunkSplitter{
		data:      data,
		chunkSize: ChunkSize,
		metadata: ChunkMetadata{
			TotalChunks: uint32((len(data) + ChunkSize - 1) / ChunkSize),
			TotalSize:   uint64(len(data)),
			Checksum:    CalculateChecksum(data),
		},
	}
}

// SetChunkSize sets custom chunk size
func (cs *ChunkSplitter) SetChunkSize(size int) {
	cs.chunkSize = size
	cs.metadata.TotalChunks = uint32((len(cs.data) + size - 1) / size)
}

// SetFileID sets the file ID for the transfer
func (cs *ChunkSplitter) SetFileID(id string) {
	cs.metadata.FileID = id
}

// Metadata returns the chunk metadata
func (cs *ChunkSplitter) Metadata() ChunkMetadata {
	return cs.metadata
}

// NextChunk returns the next chunk for upload
// Returns io.EOF when all chunks have been read
func (cs *ChunkSplitter) NextChunk() (*Chunk, error) {
	// Better implementation: track current index
	return nil, fmt.Errorf("not implemented")
}

// SplitAll splits all data into chunks at once
func (cs *ChunkSplitter) SplitAll() ([]Chunk, error) {
	var chunks []Chunk

	for i := 0; i < len(cs.data); i += cs.chunkSize {
		end := i + cs.chunkSize
		if end > len(cs.data) {
			end = len(cs.data)
		}

		chunkData := cs.data[i:end]
		chunk := Chunk{
			Index:    uint32(len(chunks)),
			Data:     chunkData,
			Checksum: CalculateChecksum(chunkData),
			Size:     uint32(len(chunkData)),
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// ChunkAssembler assembles chunks back into original data
type ChunkAssembler struct {
	metadata ChunkMetadata
	chunks   map[uint32][]byte
	received uint32
}

// NewChunkAssembler creates a new chunk assembler
func NewChunkAssembler(metadata ChunkMetadata) *ChunkAssembler {
	return &ChunkAssembler{
		metadata: metadata,
		chunks:   make(map[uint32][]byte),
		received: 0,
	}
}

// AddChunk adds a received chunk
func (ca *ChunkAssembler) AddChunk(chunk Chunk) error {
	// Verify checksum
	if !bytes.Equal(CalculateChecksum(chunk.Data), chunk.Checksum) {
		return fmt.Errorf("checksum mismatch for chunk %d", chunk.Index)
	}

	// Check if already received
	if _, exists := ca.chunks[chunk.Index]; exists {
		return nil // Already received, ignore duplicate
	}

	ca.chunks[chunk.Index] = chunk.Data
	ca.received++

	return nil
}

// IsComplete checks if all chunks have been received
func (ca *ChunkAssembler) IsComplete() bool {
	return ca.received == ca.metadata.TotalChunks
}

// Assemble assembles all chunks into complete data
func (ca *ChunkAssembler) Assemble() ([]byte, error) {
	if !ca.IsComplete() {
		return nil, fmt.Errorf("incomplete assembly: %d/%d chunks received",
			ca.received, ca.metadata.TotalChunks)
	}

	// Allocate buffer for total size
	result := make([]byte, 0, ca.metadata.TotalSize)

	// Append chunks in order
	for i := uint32(0); i < ca.metadata.TotalChunks; i++ {
		chunkData, ok := ca.chunks[i]
		if !ok {
			return nil, fmt.Errorf("missing chunk %d", i)
		}
		result = append(result, chunkData...)
	}

	// Verify final checksum
	if !bytes.Equal(CalculateChecksum(result), ca.metadata.Checksum) {
		return nil, fmt.Errorf("final checksum mismatch")
	}

	return result, nil
}

// MissingChunks returns indices of missing chunks
func (ca *ChunkAssembler) MissingChunks() []uint32 {
	var missing []uint32
	for i := uint32(0); i < ca.metadata.TotalChunks; i++ {
		if _, exists := ca.chunks[i]; !exists {
			missing = append(missing, i)
		}
	}
	return missing
}

// SerializeChunk serializes a chunk to binary format
func SerializeChunk(chunk Chunk) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write index
	if err := binary.Write(buf, binary.LittleEndian, chunk.Index); err != nil {
		return nil, err
	}

	// Write size
	if err := binary.Write(buf, binary.LittleEndian, chunk.Size); err != nil {
		return nil, err
	}

	// Write checksum length
	checksumLen := uint8(len(chunk.Checksum))
	if err := binary.Write(buf, binary.LittleEndian, checksumLen); err != nil {
		return nil, err
	}

	// Write checksum
	if _, err := buf.Write(chunk.Checksum); err != nil {
		return nil, err
	}

	// Write data
	if _, err := buf.Write(chunk.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeChunk deserializes a chunk from binary format
func DeserializeChunk(data []byte) (Chunk, error) {
	buf := bytes.NewReader(data)
	var chunk Chunk

	// Read index
	if err := binary.Read(buf, binary.LittleEndian, &chunk.Index); err != nil {
		return Chunk{}, err
	}

	// Read size
	if err := binary.Read(buf, binary.LittleEndian, &chunk.Size); err != nil {
		return Chunk{}, err
	}

	// Read checksum length
	var checksumLen uint8
	if err := binary.Read(buf, binary.LittleEndian, &checksumLen); err != nil {
		return Chunk{}, err
	}

	// Read checksum
	chunk.Checksum = make([]byte, checksumLen)
	if _, err := buf.Read(chunk.Checksum); err != nil {
		return Chunk{}, err
	}

	// Read data
	chunk.Data = make([]byte, chunk.Size)
	if _, err := buf.Read(chunk.Data); err != nil {
		return Chunk{}, err
	}

	return chunk, nil
}

// SerializeMetadata serializes metadata to binary format
func SerializeMetadata(metadata ChunkMetadata) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write total chunks
	if err := binary.Write(buf, binary.LittleEndian, metadata.TotalChunks); err != nil {
		return nil, err
	}

	// Write total size
	if err := binary.Write(buf, binary.LittleEndian, metadata.TotalSize); err != nil {
		return nil, err
	}

	// Write file ID length
	fileIDLen := uint8(len(metadata.FileID))
	if err := binary.Write(buf, binary.LittleEndian, fileIDLen); err != nil {
		return nil, err
	}

	// Write file ID
	if _, err := buf.Write([]byte(metadata.FileID)); err != nil {
		return nil, err
	}

	// Write checksum length
	checksumLen := uint8(len(metadata.Checksum))
	if err := binary.Write(buf, binary.LittleEndian, checksumLen); err != nil {
		return nil, err
	}

	// Write checksum
	if _, err := buf.Write(metadata.Checksum); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeMetadata deserializes metadata from binary format
func DeserializeMetadata(data []byte) (ChunkMetadata, error) {
	var metadata ChunkMetadata
	buf := bytes.NewReader(data)

	// Read total chunks
	if err := binary.Read(buf, binary.LittleEndian, &metadata.TotalChunks); err != nil {
		return ChunkMetadata{}, err
	}

	// Read total size
	if err := binary.Read(buf, binary.LittleEndian, &metadata.TotalSize); err != nil {
		return ChunkMetadata{}, err
	}

	// Read file ID length
	var fileIDLen uint8
	if err := binary.Read(buf, binary.LittleEndian, &fileIDLen); err != nil {
		return ChunkMetadata{}, err
	}

	// Read file ID
	fileIDBytes := make([]byte, fileIDLen)
	if _, err := buf.Read(fileIDBytes); err != nil {
		return ChunkMetadata{}, err
	}
	metadata.FileID = string(fileIDBytes)

	// Read checksum length
	var checksumLen uint8
	if err := binary.Read(buf, binary.LittleEndian, &checksumLen); err != nil {
		return ChunkMetadata{}, err
	}

	// Read checksum
	metadata.Checksum = make([]byte, checksumLen)
	if _, err := buf.Read(metadata.Checksum); err != nil {
		return ChunkMetadata{}, err
	}

	return metadata, nil
}

// StreamChunkReader reads data in chunks from a reader
type StreamChunkReader struct {
	reader    io.Reader
	chunkSize int
	index     uint32
}

// NewStreamChunkReader creates a new streaming chunk reader
func NewStreamChunkReader(reader io.Reader, chunkSize int) *StreamChunkReader {
	if chunkSize <= 0 {
		chunkSize = ChunkSize
	}
	return &StreamChunkReader{
		reader:    reader,
		chunkSize: chunkSize,
		index:     0,
	}
}

// NextChunk reads the next chunk from the stream
func (scr *StreamChunkReader) NextChunk() (*Chunk, error) {
	buf := make([]byte, scr.chunkSize)
	n, err := io.ReadFull(scr.reader, buf)

	if err == io.EOF || err == io.ErrUnexpectedEOF {
		if n == 0 {
			return nil, io.EOF
		}
		// Last partial chunk
		chunkData := buf[:n]
		return &Chunk{
			Index:    scr.index,
			Data:     chunkData,
			Checksum: CalculateChecksum(chunkData),
			Size:     uint32(n),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	chunk := &Chunk{
		Index:    scr.index,
		Data:     buf,
		Checksum: CalculateChecksum(buf),
		Size:     uint32(len(buf)),
	}
	scr.index++
	return chunk, nil
}

// StreamChunkWriter writes chunks to a writer
type StreamChunkWriter struct {
	writer    io.Writer
	metadata  ChunkMetadata
	chunks    map[uint32][]byte
	nextWrite uint32
}

// NewStreamChunkWriter creates a new streaming chunk writer
func NewStreamChunkWriter(writer io.Writer, metadata ChunkMetadata) *StreamChunkWriter {
	return &StreamChunkWriter{
		writer:    writer,
		metadata:  metadata,
		chunks:    make(map[uint32][]byte),
		nextWrite: 0,
	}
}

// AddChunk adds a chunk (may be out of order)
func (scw *StreamChunkWriter) AddChunk(chunk Chunk) error {
	if !bytes.Equal(CalculateChecksum(chunk.Data), chunk.Checksum) {
		return fmt.Errorf("checksum mismatch for chunk %d", chunk.Index)
	}

	scw.chunks[chunk.Index] = chunk.Data
	scw.writeAvailableChunks()
	return nil
}

// writeAvailableChunks writes consecutive chunks starting from nextWrite
func (scw *StreamChunkWriter) writeAvailableChunks() {
	for {
		chunkData, ok := scw.chunks[scw.nextWrite]
		if !ok {
			break
		}

		scw.writer.Write(chunkData)
		delete(scw.chunks, scw.nextWrite)
		scw.nextWrite++
	}
}

// IsComplete checks if all chunks have been written
func (scw *StreamChunkWriter) IsComplete() bool {
	return scw.nextWrite == scw.metadata.TotalChunks
}
