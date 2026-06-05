package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type TagType byte

const (
	TagEnd TagType = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
)

type Tag interface {
	Type() TagType
	String() string
}

type (
	TagByteValue struct {
		Value int8
	}
	TagShortValue struct {
		Value int16
	}
	TagIntValue struct {
		Value int32
	}
	TagLongValue struct {
		Value int64
	}
	TagFloatValue struct {
		Value float32
	}
	TagDoubleValue struct {
		Value float64
	}
	TagByteArrayValue struct {
		Value []byte
	}
	TagStringValue struct {
		Value string
	}
	TagListValue struct {
		ElementType TagType
		Elements    []Tag
	}
	TagCompoundValue struct {
		Tags map[string]Tag
	}
)

func (t *TagByteValue) Type() TagType  { return TagByte }
func (t *TagByteValue) String() string { return fmt.Sprintf("TAG_Byte: %d", t.Value) }

func (t *TagShortValue) Type() TagType  { return TagShort }
func (t *TagShortValue) String() string { return fmt.Sprintf("TAG_Short: %d", t.Value) }

func (t *TagIntValue) Type() TagType  { return TagInt }
func (t *TagIntValue) String() string { return fmt.Sprintf("TAG_Int: %d", t.Value) }

func (t *TagLongValue) Type() TagType  { return TagLong }
func (t *TagLongValue) String() string { return fmt.Sprintf("TAG_Long: %d", t.Value) }

func (t *TagFloatValue) Type() TagType  { return TagFloat }
func (t *TagFloatValue) String() string { return fmt.Sprintf("TAG_Float: %f", t.Value) }

func (t *TagDoubleValue) Type() TagType  { return TagDouble }
func (t *TagDoubleValue) String() string { return fmt.Sprintf("TAG_Double: %f", t.Value) }

func (t *TagByteArrayValue) Type() TagType { return TagByteArray }
func (t *TagByteArrayValue) String() string {
	return fmt.Sprintf("TAG_Byte_Array: %d bytes", len(t.Value))
}

func (t *TagStringValue) Type() TagType  { return TagString }
func (t *TagStringValue) String() string { return fmt.Sprintf("TAG_String: %s", t.Value) }

func (t *TagListValue) Type() TagType { return TagDouble }
func (t *TagListValue) String() string {
	return fmt.Sprintf("TAG_List<%d>: %d elements", t.ElementType, len(t.Elements))
}

func (t *TagCompoundValue) Type() TagType { return TagCompound }
func (t *TagCompoundValue) String() string {
	return fmt.Sprintf("TAG_Compound: %d tags", len(t.Tags))
}

type NBTReader struct {
	r   io.Reader
	err error
}

func NewNBTReader(r io.Reader) *NBTReader {
	return &NBTReader{r: r}
}

func (nr *NBTReader) readByte() byte {
	if nr.err != nil {
		return 0
	}
	b := make([]byte, 1)
	_, nr.err = io.ReadFull(nr.r, b)
	return b[0]
}

func (nr *NBTReader) readBytes(n int) []byte {
	if nr.err != nil {
		return nil
	}
	b := make([]byte, n)
	_, nr.err = io.ReadFull(nr.r, b)
	return b
}

func (nr *NBTReader) readShort() int16 {
	if nr.err != nil {
		return 0
	}

	b := nr.readBytes(2)
	if nr.err != nil {
		return 0
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (nr *NBTReader) readInt() int32 {
	if nr.err != nil {
		return 0
	}

	b := nr.readBytes(4)
	if nr.err != nil {
		return 0
	}
	return int32(binary.BigEndian.Uint32(b))
}

func (nr *NBTReader) readLong() int64 {
	if nr.err != nil {
		return 0
	}

	b := nr.readBytes(8)
	if nr.err != nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}

func (nr *NBTReader) readFloat() float32 {
	if nr.err != nil {
		return 0
	}
	return math.Float32frombits(binary.BigEndian.Uint32(nr.readBytes(4)))
}

func (nr *NBTReader) readDouble() float64 {
	if nr.err != nil {
		return 0
	}
	return math.Float64frombits(binary.BigEndian.Uint64(nr.readBytes(8)))
}

func (nr *NBTReader) readString() string {
	if nr.err != nil {
		return ""
	}
	length := nr.readShort()
	if nr.err != nil {
		return ""
	}
	return string(nr.readBytes(int(length)))
}

func (nr *NBTReader) ReadTag() (name string, tag Tag, err error) {
	tagType := TagType(nr.readByte())
	if nr.err != nil {
		return "", nil, nr.err
	}

	if tagType == TagEnd {
		return "", nil, nil
	}

	name = nr.readString()
	if nr.err != nil {
		return "", nil, nr.err
	}

	tag = nr.readTagValue(tagType)
	if nr.err != nil {
		return "", nil, nr.err
	}

	return name, tag, nil
}

func (nr *NBTReader) readTagValue(tagType TagType) Tag {
	if nr.err != nil {
		return nil
	}

	switch tagType {
	case TagByte:
		return &TagByteValue{Value: int8(nr.readByte())}

	case TagShort:
		return &TagShortValue{Value: nr.readShort()}

	case TagInt:
		return &TagIntValue{Value: nr.readInt()}

	case TagLong:
		return &TagLongValue{Value: nr.readLong()}

	case TagFloat:
		return &TagFloatValue{Value: nr.readFloat()}

	case TagDouble:
		return &TagDoubleValue{Value: nr.readDouble()}

	case TagByteArray:
		length := nr.readInt()
		if nr.err != nil {
			return nil
		}
		return &TagByteArrayValue{Value: nr.readBytes(int(length))}

	case TagString:
		return &TagStringValue{Value: nr.readString()}

	case TagList:
		elementType := TagType(nr.readByte())
		length := nr.readInt()
		if nr.err != nil {
			return nil
		}

		list := &TagListValue{
			ElementType: elementType,
			Elements:    make([]Tag, length),
		}

		for i := 0; i < int(length); i++ {
			list.Elements[i] = nr.readTagValue(elementType)
			if nr.err != nil {
				return nil
			}
		}
		return list

	case TagCompound:
		compound := &TagCompoundValue{Tags: make(map[string]Tag)}
		for {
			name, tag, err := nr.ReadTag()
			if err != nil {
				nr.err = err
				return nil
			}
			if tag == nil {
				break
			}
			compound.Tags[name] = tag
		}
		return compound

	default:
		nr.err = errors.New("unknown tag type")
		return nil
	}
}

func (nr *NBTReader) Error() error {
	return nr.err
}

// ===== NBT Writer =====

type NBTWriter struct {
	w   io.Writer
	err error
}

func NewNBTWriter(w io.Writer) *NBTWriter {
	return &NBTWriter{w: w}
}

func (nw *NBTWriter) writeByte(b byte) {
	if nw.err != nil {
		return
	}
	_, nw.err = nw.w.Write([]byte{b})
}

func (nw *NBTWriter) writeBytes(b []byte) {
	if nw.err != nil {
		return
	}
	_, nw.err = nw.w.Write(b)
}

func (nw *NBTWriter) writeShort(v int16) {
	if nw.err != nil {
		return
	}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	nw.writeBytes(b)
}

func (nw *NBTWriter) writeInt(v int32) {
	if nw.err != nil {
		return
	}
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	nw.writeBytes(b)
}

func (nw *NBTWriter) writeLong(v int64) {
	if nw.err != nil {
		return
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	nw.writeBytes(b)
}

func (nw *NBTWriter) writeFloat(v float32) {
	if nw.err != nil {
		return
	}
	nw.writeInt(int32(math.Float32bits(v)))
}

func (nw *NBTWriter) writeDouble(v float64) {
	if nw.err != nil {
		return
	}
	nw.writeLong(int64(math.Float64bits(v)))
}

func (nw *NBTWriter) writeString(v string) {
	if nw.err != nil {
		return
	}
	nw.writeShort(int16(len(v)))
	nw.writeBytes([]byte(v))
}

func (nw *NBTWriter) WriteTag(name string, tag Tag) {
	if nw.err != nil {
		return
	}

	nw.writeByte(byte(tag.Type()))
	nw.writeString(name)
	nw.writeTagValue(tag)
}

func (nw *NBTWriter) writeTagValue(tag Tag) {
	if nw.err != nil {
		return
	}

	switch t := tag.(type) {
	case *TagByteValue:
		nw.writeByte(byte(t.Value))

	case *TagShortValue:
		nw.writeShort(t.Value)

	case *TagIntValue:
		nw.writeInt(t.Value)

	case *TagLongValue:
		nw.writeLong(t.Value)

	case *TagFloatValue:
		nw.writeFloat(t.Value)

	case *TagDoubleValue:
		nw.writeDouble(t.Value)

	case *TagByteArrayValue:
		nw.writeInt(int32(len(t.Value)))
		nw.writeBytes(t.Value)

	case *TagStringValue:
		nw.writeString(t.Value)

	case *TagListValue:
		nw.writeByte(byte(t.ElementType))
		nw.writeInt(int32(len(t.Elements)))
		for _, elem := range t.Elements {
			nw.writeTagValue(elem)
		}

	case *TagCompoundValue:
		for name, tag := range t.Tags {
			nw.WriteTag(name, tag)
		}
		nw.writeByte(byte(TagEnd))

	default:
		nw.err = errors.New("unknown tag type")
	}
}

func (nw *NBTWriter) Error() error {
	return nw.err
}
