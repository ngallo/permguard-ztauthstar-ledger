// Copyright 2024 Nitro Agility S.r.l.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package objects

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
)

// SerializeBlob serializes a blob object.
func (m *ObjectManager) SerializeBlob(header *ObjectHeader, data []byte) ([]byte, error) {
	if header == nil {
		return nil, errors.New("objects: header is nil")
	}

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, header.isNativeLanguage); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, header.languageID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, header.languageVersionID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, header.languageTypeID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, header.codeTypeID); err != nil {
		return nil, err
	}
	encodedCodeID := base64.StdEncoding.EncodeToString([]byte(header.codeID))
	codeIDBytes := []byte(encodedCodeID)
	codeIDBytesLength := uint16(len(codeIDBytes))
	if err := binary.Write(&buffer, binary.BigEndian, codeIDBytesLength); err != nil {
		return nil, err
	}
	if _, err := buffer.Write(codeIDBytes); err != nil {
		return nil, err
	}
	if err := buffer.WriteByte(PacketNullByte); err != nil {
		return nil, err
	}
	content := append(buffer.Bytes(), data...)

	return content, nil
}

// DeserializeBlob deserializes a blob object.
func (m *ObjectManager) DeserializeBlob(data []byte) (*ObjectHeader, []byte, error) {
	if len(data) < 1 {
		return nil, nil, errors.New("objects: data is too short to contain an ObjectHeader")
	}

	delimiterIndex := bytes.IndexByte(data, PacketNullByte)
	if delimiterIndex == -1 {
		return nil, nil, errors.New("objects: null packet delimiter not found")
	}
	if delimiterIndex < 13 {
		return nil, nil, errors.New("objects: data is too short to contain a complete ObjectHeader")
	}

	header := &ObjectHeader{}
	if err := binary.Read(bytes.NewReader(data[:delimiterIndex]), binary.BigEndian, &header.isNativeLanguage); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(bytes.NewReader(data[1:delimiterIndex]), binary.BigEndian, &header.languageID); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(bytes.NewReader(data[5:delimiterIndex]), binary.BigEndian, &header.languageVersionID); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(bytes.NewReader(data[9:delimiterIndex]), binary.BigEndian, &header.languageTypeID); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(bytes.NewReader(data[13:delimiterIndex]), binary.BigEndian, &header.codeTypeID); err != nil {
		return nil, nil, err
	}
	var length uint16
	if err := binary.Read(bytes.NewReader(data[17:19]), binary.BigEndian, &length); err != nil {
		return nil, nil, errors.New("objects: failed to read codeID length")
	}
	encodedCodeIDBytes := data[19 : 19+length]
	header.codeID = string(encodedCodeIDBytes)
	decodedCodeID, err := base64.StdEncoding.DecodeString(header.codeID)
	if err != nil {
		return nil, nil, errors.New("objects: failed to decode codeID")
	}
	header.codeID = string(decodedCodeID)
	remainingData := data[delimiterIndex+1:]
	return header, remainingData, nil
}
