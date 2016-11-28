/*
 Copyright 2016 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package proxy

import (
	"bytes"
	"log"
	"strings"
)

var WRITE_COMMANDS = []string{"INSERT",
	"DELETE", "UPSERT", "UPDATE", "CREATE",
	"DROP", "ALTER", "COPY"}

func IsWrite(buf []byte) bool {
	var msgLen int32
	var query string
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	query = string(buf[5:msgLen])
	log.Printf("IsWrite: msglen=%d query=%s\n", msgLen, query)
	upperQuery := strings.ToUpper(query)

	for i := range WRITE_COMMANDS {
		if strings.Contains(upperQuery, WRITE_COMMANDS[i]) {
			log.Println(WRITE_COMMANDS[i] + " was parsed out of query")
			return true
		}
	}

	return false
}

var START = []byte{'/', '*'}
var END = []byte{'*', '/'}

//the annotation approach
//assume a write if there is no comment in the SQL
//or if there are no keywords in the comment
func IsWriteAnno(buf []byte) bool {
	var msgLen int32
	var query string
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	query = string(buf[5:msgLen])
	log.Printf("IsWrite: msglen=%d query=%s\n", msgLen, query)

	querybuf := buf[5:msgLen]
	startPos := bytes.Index(buf, START)
	endPos := bytes.Index(buf, END)
	if startPos < 0 || endPos < 0 {
		log.Println("no comment found..assuming write case")
		return true
	}
	startPos = startPos + 5 //add 5 for msg header length
	endPos = endPos + 5     //add 5 for msg header length

	comment := buf[bytes.Index(querybuf, START)+2+5 : bytes.Index(querybuf, END)+5]
	log.Printf("comment=[%s]\n", string(comment))

	keywords := bytes.Split(comment, []byte(","))
	var stateful = false
	var write = true
	var keywordFound = false
	for i := 0; i < len(keywords); i++ {
		log.Printf("keyword=[%s]\n", string(bytes.TrimSpace(keywords[i])))
		if string(bytes.TrimSpace(keywords[i])) == "read" {
			log.Println("read was found")
			write = false
			keywordFound = true
		}
		if string(bytes.TrimSpace(keywords[i])) == "write" {
			log.Println("write was found")
			write = true
			keywordFound = true
		}
		if string(bytes.TrimSpace(keywords[i])) == "stateful" {
			log.Println("stateful was found")
			stateful = true
			keywordFound = true
		}
	}
	log.Printf("write=%t stateful=%t\n", write, stateful)
	if keywordFound == false {
		log.Println("no keywords found in SQL comment..assuming write")
	}

	return write
}