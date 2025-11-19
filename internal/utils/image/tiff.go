package image

import (
	"kano/internal/utils/numbers"
)

func processMSB(b []byte, useMSB bool) []byte {
	if useMSB {
		c := make([]byte, len(b))
		copy(c, b)
		for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
			c[i], c[j] = c[j], c[i]
		}
		return c
	}

	return b
}

func BuildTIFF(ifds []IFD, useMSB bool) ([]byte, error) {
	result := []byte{}

	// TIFF Headers
	if useMSB {
		result = append(result, 0x4d, 0x4d)
	} else {
		result = append(result, 0x49, 0x49)
	}
	result = append(result, processMSB([]byte{42, 0}, useMSB)...)      // The carefully chosen arbitrary number (42)
	result = append(result, processMSB([]byte{8, 0, 0, 0}, useMSB)...) // In this case, IFD0 always start at offset 8

	if len(ifds) == 0 {
		return nil, ErrNoIFDProvided
	}

	// Process all ifds
	for ifdIdx, ifd := range ifds {
		ifdStartOffset := len(result)

		numberOfEntries := processMSB(numbers.Int16ToByteLSB(int(ifd.NumberOfEntries)), useMSB)
		result = append(result, numberOfEntries...)

		if int(ifd.NumberOfEntries) != len(ifd.Entries) {
			return nil, ErrMissmatchIFDSize
		}

		ifdEndOffset := ifdStartOffset + 2 + int(ifd.NumberOfEntries)*12 + 4 // Start + number of entries + entries size + offset next IFD
		valueToAppend := []byte{}                                            // Any values that cannot fit in the entry value (4-byte) will be placed after the end of this IFD

		// Process ifd's entries
		for _, entry := range ifd.Entries {
			entryByte := []byte{}

			entryTag := processMSB(numbers.Int16ToByteLSB(int(entry.Tag)), useMSB)
			entryType := processMSB(numbers.Int16ToByteLSB(int(entry.Type)), useMSB)
			entryCount := processMSB(numbers.Int32ToByteLSB(int(entry.Count)), useMSB)
			entryValue := entry.Value

			entryTypeSize := ENTRY_TYPE_SIZE[entry.Type]
			entrySize := entryTypeSize * int(entry.Count)
			if len(entryValue) != entrySize {
				return nil, ErrMissmatchEntrySize
			}

			// I am going delusional, let's just keep it
			// Somehow it works as intended
			actualValue := []byte{}
			if useMSB && entryTypeSize != 1 {
				mult := 1
				if entryTypeSize == 8 {
					mult = 2
				}
				for i := range int(entry.Count) * mult {
					idx := int(i)
					start, end := entryTypeSize/mult*idx, entryTypeSize/mult*(idx+1)
					individualValue := processMSB(entry.Value[start:end], true)
					actualValue = append(actualValue, individualValue...)
				}
				entryValue = actualValue
			}

			if ENTRY_TYPE_SIZE[entry.Type]*int(entry.Count) > 4 {
				if len(entryValue)%2 == 1 {
					entryValue = append(entryValue, 0) // Keeping it on the word boundary
				}
				valueToAppend = append(valueToAppend, entryValue...)

				offset := ifdEndOffset
				ifdEndOffset += entrySize
				entryValue = processMSB(numbers.Int32ToByteLSB(offset), useMSB)

			}

			if len(entryValue) < 4 {
				for range 4 - len(entryValue) {
					entryValue = append(entryValue, 0) // Left justified
				}
			}

			entryByte = append(entryByte, entryTag...)
			entryByte = append(entryByte, entryType...)
			entryByte = append(entryByte, entryCount...)
			entryByte = append(entryByte, entryValue...)

			result = append(result, entryByte...)
		}

		// Yeah, the next IFD
		if ifdIdx != len(ifds)-1 {
			result = append(result, numbers.Int32ToByteLSB(ifdEndOffset)...)
		} else {
			result = append(result, []byte{0, 0, 0, 0}...)
		}

		if len(valueToAppend)%2 == 1 {
			panic("PANIK! there is a data with odd length >:(")
		}

		if len(valueToAppend) > 0 {
			result = append(result, valueToAppend...)
		}
	}

	return result, nil
}
