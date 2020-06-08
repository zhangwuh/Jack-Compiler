// core_test.go  * Created on  2020/6/8
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
	"fmt"
	"testing"
)

func TestNonTerminalToken_AsText(t *testing.T) {
	token := &NonTerminalToken{
		tokenType: Class,
		tokens: []Token{
			&TerminalToken{
				tokenType: Keyword,
				val: "function",
			},
			&NonTerminalToken{
				tokenType: VarDec,
				tokens: []Token{
					&TerminalToken{
						tokenType: Keyword,
						val: "var",
					},
					&TerminalToken{
						tokenType: Keyword,
						val: "int",
					},
					&TerminalToken{
						tokenType: Identifier,
						val: "length",
					},
					&TerminalToken{
						tokenType: Symbol,
						val: ";",
					},
				},
			},
		},
	}

	fmt.Println(token.AsText())
}
