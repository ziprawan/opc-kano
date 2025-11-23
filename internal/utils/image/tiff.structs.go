package image

type EntryType = uint8
type ImageFileDirectory struct {
	NumberOfEntries uint16
	Entries         []DirectoryEntry
}
type DirectoryEntry struct {
	Tag   uint16
	Type  EntryType
	Count uint32
	Value []byte
}

type IFD = ImageFileDirectory
type DE = DirectoryEntry

const (
	_ EntryType = iota
	EntryTypeByte
	EntryTypeASCII
	EntryTypeShort
	EntryTypeLong
	EntryTypeRational
	EntryTypeSByte
	EntryTypeUndefined
	EntryTypeSShort
	EntryTypeSLong
	EntryTypeSRational
	EntryTypeFloat
	EntryTypeDouble
)

var ENTRY_TYPE_SIZE = map[EntryType]int{
	EntryTypeByte:      1,
	EntryTypeASCII:     1,
	EntryTypeShort:     2,
	EntryTypeLong:      4,
	EntryTypeRational:  8,
	EntryTypeSByte:     1,
	EntryTypeUndefined: 1,
	EntryTypeSShort:    2,
	EntryTypeSLong:     4,
	EntryTypeSRational: 8,
	EntryTypeFloat:     4,
	EntryTypeDouble:    8,
}
