// tokenizer.go  * Created on  2020/6/8
// Copyright (c) 2020 YueTu
// YueTu TECHNOLOGY CO.,LTD. All Rights Reserved.
//
// This software is the confidential and proprietary information of
// YueTu Ltd. ("Confidential Information").
// You shall not disclose such Confidential Information and shall use
// it only in accordance with the terms of the license agreement you
// entered into with YueTu Ltd.

package jack_compiler

import (
	"bufio"
	"fmt"
	"io"
)

type tokenizer struct {
}

func (tokenizer *tokenizer) Tokenize(rd io.Reader) (tokens []TerminalToken) {
	reader := bufio.NewReader(rd)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(fmt.Sprintf("read from reader err:%s", err.Error()))
			return
		}
		tokens = append(tokens, tokenize(line))
	}
}

func tokenize(bs []byte) TerminalToken {

}
