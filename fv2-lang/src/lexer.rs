// FV 2.0 Lexer: V-호환 문법 지원
// V 언어의 토큰을 정의하고, 소스 코드를 토큰 스트림으로 변환

use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Token {
    // 키워드
    Fn,
    Let,
    Mut,
    Const,
    If,
    Else,
    For,
    In,
    Match,
    Type,
    Struct,
    Interface,
    Enum,
    Trait,
    Impl,
    Return,
    Module,
    Import,

    // 리터럴
    Identifier(String),
    Integer(i64),
    Float(f64),
    String(String),
    RawString(String),
    True,
    False,
    None,

    // 연산자
    Plus,           // +
    Minus,          // -
    Star,           // *
    Slash,          // /
    Percent,        // %
    Caret,          // ^
    Ampersand,      // &
    Pipe,           // |
    Tilde,          // ~
    LeftShift,      // <<
    RightShift,     // >>

    // 할당 & 비교
    Assign,         // =
    PlusAssign,     // +=
    MinusAssign,    // -=
    StarAssign,     // *=
    SlashAssign,    // /=
    ColonAssign,    // :=
    Eq,             // ==
    Ne,             // !=
    Lt,             // <
    Le,             // <=
    Gt,             // >
    Ge,             // >=
    LogicalAnd,     // &&
    LogicalOr,      // ||
    Not,            // !
    Question,       // ?
    Dot,            // .
    DoubleDot,      // ..
    DotDotEq,       // ..=

    // 화살표 & 복합
    Arrow,          // ->
    FatArrow,       // =>
    Colon,          // :
    DoubleColon,    // ::

    // 구분자
    LParen,         // (
    RParen,         // )
    LBrace,         // {
    RBrace,         // }
    LBracket,       // [
    RBracket,       // ]
    Comma,          // ,
    Semicolon,      // ;
    At,             // @

    // 특수
    Eof,
}

#[derive(Debug, Clone)]
pub struct TokenInfo {
    pub token: Token,
    pub line: usize,
    pub column: usize,
    pub text: String,
}

pub struct Lexer {
    input: Vec<char>,
    pos: usize,
    line: usize,
    column: usize,
    keywords: HashMap<String, Token>,
}

impl Lexer {
    pub fn new(input: &str) -> Result<Self, String> {
        // 보안: 입력 크기 제한
        if input.len() > 10_000_000 {
            return Err(format!("Input too large: {} bytes", input.len()));
        }

        // 보안: NULL 바이트 확인
        if input.contains('\0') {
            return Err("Input contains null bytes".to_string());
        }

        let mut keywords = HashMap::new();
        keywords.insert("fn".to_string(), Token::Fn);
        keywords.insert("let".to_string(), Token::Let);
        keywords.insert("mut".to_string(), Token::Mut);
        keywords.insert("const".to_string(), Token::Const);
        keywords.insert("if".to_string(), Token::If);
        keywords.insert("else".to_string(), Token::Else);
        keywords.insert("for".to_string(), Token::For);
        keywords.insert("in".to_string(), Token::In);
        keywords.insert("match".to_string(), Token::Match);
        keywords.insert("type".to_string(), Token::Type);
        keywords.insert("struct".to_string(), Token::Struct);
        keywords.insert("interface".to_string(), Token::Interface);
        keywords.insert("enum".to_string(), Token::Enum);
        keywords.insert("trait".to_string(), Token::Trait);
        keywords.insert("impl".to_string(), Token::Impl);
        keywords.insert("return".to_string(), Token::Return);
        keywords.insert("module".to_string(), Token::Module);
        keywords.insert("import".to_string(), Token::Import);
        keywords.insert("true".to_string(), Token::True);
        keywords.insert("false".to_string(), Token::False);
        keywords.insert("none".to_string(), Token::None);

        Ok(Lexer {
            input: input.chars().collect(),
            pos: 0,
            line: 1,
            column: 1,
            keywords,
        })
    }

    /// 모든 토큰 수집
    pub fn tokenize(mut self) -> Result<Vec<TokenInfo>, String> {
        let mut tokens = Vec::new();

        loop {
            self.skip_whitespace_and_comments();

            if self.is_at_end() {
                tokens.push(TokenInfo {
                    token: Token::Eof,
                    line: self.line,
                    column: self.column,
                    text: String::new(),
                });
                break;
            }

            match self.next_token() {
                Ok(token_info) => tokens.push(token_info),
                Err(e) => return Err(format!(
                    "Lexer error at line {}, column {}: {}",
                    self.line, self.column, e
                )),
            }
        }

        Ok(tokens)
    }

    fn next_token(&mut self) -> Result<TokenInfo, String> {
        let start_line = self.line;
        let start_column = self.column;
        let start_pos = self.pos;

        let ch = self.current_char();

        let token = match ch {
            '(' => {
                self.advance();
                Token::LParen
            }
            ')' => {
                self.advance();
                Token::RParen
            }
            '{' => {
                self.advance();
                Token::LBrace
            }
            '}' => {
                self.advance();
                Token::RBrace
            }
            '[' => {
                self.advance();
                Token::LBracket
            }
            ']' => {
                self.advance();
                Token::RBracket
            }
            ',' => {
                self.advance();
                Token::Comma
            }
            ';' => {
                self.advance();
                Token::Semicolon
            }
            '@' => {
                self.advance();
                Token::At
            }
            '~' => {
                self.advance();
                Token::Tilde
            }
            '+' => {
                self.advance();
                if self.current_char() == '=' {
                    self.advance();
                    Token::PlusAssign
                } else {
                    Token::Plus
                }
            }
            '-' => {
                self.advance();
                match self.current_char() {
                    '=' => {
                        self.advance();
                        Token::MinusAssign
                    }
                    '>' => {
                        self.advance();
                        Token::Arrow
                    }
                    _ => Token::Minus,
                }
            }
            '*' => {
                self.advance();
                if self.current_char() == '=' {
                    self.advance();
                    Token::StarAssign
                } else {
                    Token::Star
                }
            }
            '/' => {
                self.advance();
                if self.current_char() == '=' {
                    self.advance();
                    Token::SlashAssign
                } else {
                    Token::Slash
                }
            }
            '%' => {
                self.advance();
                Token::Percent
            }
            '^' => {
                self.advance();
                Token::Caret
            }
            '&' => {
                self.advance();
                if self.current_char() == '&' {
                    self.advance();
                    Token::LogicalAnd
                } else {
                    Token::Ampersand
                }
            }
            '|' => {
                self.advance();
                if self.current_char() == '|' {
                    self.advance();
                    Token::LogicalOr
                } else {
                    Token::Pipe
                }
            }
            '!' => {
                self.advance();
                if self.current_char() == '=' {
                    self.advance();
                    Token::Ne
                } else {
                    Token::Not
                }
            }
            '?' => {
                self.advance();
                Token::Question
            }
            '=' => {
                self.advance();
                match self.current_char() {
                    '=' => {
                        self.advance();
                        Token::Eq
                    }
                    '>' => {
                        self.advance();
                        Token::FatArrow
                    }
                    _ => Token::Assign,
                }
            }
            '<' => {
                self.advance();
                match self.current_char() {
                    '=' => {
                        self.advance();
                        Token::Le
                    }
                    '<' => {
                        self.advance();
                        Token::LeftShift
                    }
                    _ => Token::Lt,
                }
            }
            '>' => {
                self.advance();
                match self.current_char() {
                    '=' => {
                        self.advance();
                        Token::Ge
                    }
                    '>' => {
                        self.advance();
                        Token::RightShift
                    }
                    _ => Token::Gt,
                }
            }
            ':' => {
                self.advance();
                match self.current_char() {
                    ':' => {
                        self.advance();
                        Token::DoubleColon
                    }
                    '=' => {
                        self.advance();
                        Token::ColonAssign
                    }
                    _ => Token::Colon,
                }
            }
            '.' => {
                self.advance();
                match self.current_char() {
                    '.' => {
                        self.advance();
                        if self.current_char() == '=' {
                            self.advance();
                            Token::DotDotEq
                        } else {
                            Token::DoubleDot
                        }
                    }
                    _ => Token::Dot,
                }
            }
            '"' => {
                self.advance();
                self.read_string('"')?
            }
            '\'' => {
                self.advance();
                self.read_string('\'')?
            }
            '`' => {
                self.advance();
                self.read_raw_string()?
            }
            _ if ch.is_ascii_digit() => self.read_number()?,
            _ if ch.is_alphabetic() || ch == '_' => self.read_identifier(),
            _ => return Err(format!("Unexpected character: '{}'", ch)),
        };

        let text = self.input[start_pos..self.pos].iter().collect();

        Ok(TokenInfo {
            token,
            line: start_line,
            column: start_column,
            text,
        })
    }

    fn read_identifier(&mut self) -> Token {
        let start = self.pos;

        while !self.is_at_end() && (self.current_char().is_alphanumeric() || self.current_char() == '_') {
            self.advance();
        }

        let ident: String = self.input[start..self.pos].iter().collect();

        self.keywords
            .get(&ident)
            .cloned()
            .unwrap_or(Token::Identifier(ident))
    }

    fn read_number(&mut self) -> Result<Token, String> {
        let start = self.pos;

        // 정수 부분
        while !self.is_at_end() && self.current_char().is_ascii_digit() {
            self.advance();
        }

        // 부동소수점 확인
        if self.current_char() == '.' && self.peek_next().is_ascii_digit() {
            self.advance(); // '.' 건너뛰기

            while !self.is_at_end() && self.current_char().is_ascii_digit() {
                self.advance();
            }

            let num_str: String = self.input[start..self.pos].iter().collect();
            let value = num_str.parse::<f64>()
                .map_err(|_| format!("Invalid float: {}", num_str))?;
            return Ok(Token::Float(value));
        }

        let num_str: String = self.input[start..self.pos].iter().collect();
        let value = num_str.parse::<i64>()
            .map_err(|_| format!("Invalid integer: {}", num_str))?;
        Ok(Token::Integer(value))
    }

    fn read_string(&mut self, quote: char) -> Result<Token, String> {
        let start = self.pos;

        while !self.is_at_end() && self.current_char() != quote {
            if self.current_char() == '\\' {
                self.advance();
                if !self.is_at_end() {
                    self.advance();
                }
            } else {
                if self.current_char() == '\n' {
                    self.line += 1;
                    self.column = 1;
                } else {
                    self.column += 1;
                }
                self.advance();
            }
        }

        if self.is_at_end() {
            return Err("Unterminated string".to_string());
        }

        let content: String = self.input[start..self.pos].iter().collect();
        self.advance(); // 닫는 따옴표 건너뛰기

        Ok(Token::String(self.unescape_string(&content)))
    }

    fn read_raw_string(&mut self) -> Result<Token, String> {
        let start = self.pos;

        while !self.is_at_end() && self.current_char() != '`' {
            if self.current_char() == '\n' {
                self.line += 1;
                self.column = 1;
            }
            self.advance();
        }

        if self.is_at_end() {
            return Err("Unterminated raw string".to_string());
        }

        let content: String = self.input[start..self.pos].iter().collect();
        self.advance(); // 닫는 백틱 건너뛰기

        Ok(Token::RawString(content))
    }

    fn unescape_string(&self, s: &str) -> String {
        let mut result = String::new();
        let mut chars = s.chars();

        while let Some(ch) = chars.next() {
            if ch == '\\' {
                match chars.next() {
                    Some('n') => result.push('\n'),
                    Some('t') => result.push('\t'),
                    Some('r') => result.push('\r'),
                    Some('\\') => result.push('\\'),
                    Some('"') => result.push('"'),
                    Some('\'') => result.push('\''),
                    Some(c) => result.push(c),
                    None => result.push('\\'),
                }
            } else {
                result.push(ch);
            }
        }

        result
    }

    fn skip_whitespace_and_comments(&mut self) {
        while !self.is_at_end() {
            match self.current_char() {
                ' ' | '\t' | '\r' => {
                    self.advance();
                }
                '\n' => {
                    self.line += 1;
                    self.column = 0;
                    self.advance();
                }
                '/' if self.peek_next() == '/' => {
                    // 한 줄 주석
                    while !self.is_at_end() && self.current_char() != '\n' {
                        self.advance();
                    }
                }
                '/' if self.peek_next() == '*' => {
                    // 블록 주석
                    self.advance();
                    self.advance();
                    while !self.is_at_end() {
                        if self.current_char() == '*' && self.peek_next() == '/' {
                            self.advance();
                            self.advance();
                            break;
                        }
                        if self.current_char() == '\n' {
                            self.line += 1;
                            self.column = 0;
                        }
                        self.advance();
                    }
                }
                _ => break,
            }
        }
    }

    fn current_char(&self) -> char {
        if self.pos < self.input.len() {
            self.input[self.pos]
        } else {
            '\0'
        }
    }

    fn peek_next(&self) -> char {
        if self.pos + 1 < self.input.len() {
            self.input[self.pos + 1]
        } else {
            '\0'
        }
    }

    fn advance(&mut self) {
        if !self.is_at_end() {
            self.pos += 1;
            self.column += 1;
        }
    }

    fn is_at_end(&self) -> bool {
        self.pos >= self.input.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_basic_tokens() {
        let code = "fn main() { let x = 5; }";
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[0].token, Token::Fn);
        assert_eq!(tokens[1].token, Token::Identifier("main".to_string()));
        assert_eq!(tokens[2].token, Token::LParen);
        assert_eq!(tokens[3].token, Token::RParen);
        assert_eq!(tokens[4].token, Token::LBrace);
    }

    #[test]
    fn test_number_literal() {
        let code = "42 3.14";
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[0].token, Token::Integer(42));
        assert_eq!(tokens[1].token, Token::Float(3.14));
    }

    #[test]
    fn test_string_literal() {
        let code = r#""hello" 'world'"#;
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[0].token, Token::String("hello".to_string()));
        assert_eq!(tokens[1].token, Token::String("world".to_string()));
    }

    #[test]
    fn test_operators() {
        let code = "a + b - c * d / e";
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[1].token, Token::Plus);
        assert_eq!(tokens[3].token, Token::Minus);
        assert_eq!(tokens[5].token, Token::Star);
        assert_eq!(tokens[7].token, Token::Slash);
    }

    #[test]
    fn test_colon_assign() {
        let code = "x := 10";
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[0].token, Token::Identifier("x".to_string()));
        assert_eq!(tokens[1].token, Token::ColonAssign);
        assert_eq!(tokens[2].token, Token::Integer(10));
    }

    #[test]
    fn test_comments() {
        let code = "// comment\nlet x = 5; /* block */";
        let lexer = Lexer::new(code).unwrap();
        let tokens = lexer.tokenize().unwrap();

        assert_eq!(tokens[0].token, Token::Let);
    }
}
