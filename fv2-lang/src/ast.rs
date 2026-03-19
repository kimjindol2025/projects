// FV 2.0 AST: Abstract Syntax Tree 정의
// V 언어 구문을 트리 구조로 표현

#[derive(Debug, Clone, PartialEq)]
pub struct Program {
    pub definitions: Vec<Definition>,
    pub main_body: Vec<Statement>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Definition {
    Function {
        name: String,
        params: Vec<(String, Type)>,
        return_type: Option<Type>,
        body: Vec<Statement>,
    },
    Type {
        name: String,
        fields: Vec<(String, Type)>,
    },
    Struct {
        name: String,
        fields: Vec<(String, Type, bool)>, // (name, type, is_mutable)
    },
    Interface {
        name: String,
        methods: Vec<(String, Vec<Type>, Option<Type>)>, // (name, params, return_type)
    },
    Enum {
        name: String,
        variants: Vec<(String, Option<Vec<Type>>)>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum Statement {
    Let {
        name: String,
        typ: Option<Type>,
        init: Option<Expression>,
        mutable: bool,
    },
    Const {
        name: String,
        typ: Option<Type>,
        value: Expression,
    },
    If {
        cond: Expression,
        then_body: Vec<Statement>,
        else_body: Option<Vec<Statement>>,
    },
    For {
        var: String,
        iter: Expression,
        body: Vec<Statement>,
    },
    ForRange {
        var: String,
        start: Expression,
        end: Expression,
        body: Vec<Statement>,
    },
    Match {
        expr: Expression,
        arms: Vec<(Pattern, Vec<Statement>)>,
    },
    Expr(Expression),
    Return(Option<Expression>),
    Break,
    Continue,
    Block(Vec<Statement>),
}

#[derive(Debug, Clone, PartialEq)]
pub enum Pattern {
    Literal(Expression),
    Identifier(String),
    Variant {
        name: String,
        fields: Option<Vec<String>>,
    },
    Wild,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Expression {
    // 리터럴
    Integer(i64),
    Float(f64),
    String(String),
    Bool(bool),
    None,

    // 변수
    Identifier(String),

    // 연산
    Binary {
        left: Box<Expression>,
        op: BinaryOp,
        right: Box<Expression>,
    },
    Unary {
        op: UnaryOp,
        operand: Box<Expression>,
    },

    // 함수/메서드 호출
    Call {
        func: Box<Expression>,
        args: Vec<Expression>,
    },
    Method {
        object: Box<Expression>,
        method: String,
        args: Vec<Expression>,
    },

    // 접근
    Index {
        object: Box<Expression>,
        index: Box<Expression>,
    },
    Field {
        object: Box<Expression>,
        field: String,
    },

    // 제어
    If {
        cond: Box<Expression>,
        then_expr: Box<Expression>,
        else_expr: Option<Box<Expression>>,
    },
    Match {
        expr: Box<Expression>,
        arms: Vec<(Pattern, Expression)>,
    },

    // 컬렉션
    Array(Vec<Expression>),
    Struct {
        name: String,
        fields: Vec<(String, Expression)>,
    },

    // 타입 변환
    Cast {
        expr: Box<Expression>,
        target_type: Type,
    },

    // 에러 처리
    ErrorPropagation(Box<Expression>), // ?
    Result {
        ok: Box<Expression>,
        err: Option<Box<Expression>>,
    },

    // Block
    Block(Vec<Statement>, Option<Box<Expression>>),
}

#[derive(Debug, Clone, PartialEq)]
pub enum BinaryOp {
    // 산술
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Pow,

    // 비트
    BitAnd,
    BitOr,
    BitXor,
    LeftShift,
    RightShift,

    // 비교
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,

    // 논리
    And,
    Or,

    // 할당
    Assign,
    AddAssign,
    SubAssign,
    MulAssign,
    DivAssign,

    // 기타
    Range,      // ..
    RangeInc,   // ..=
    As,         // as
}

#[derive(Debug, Clone, PartialEq)]
pub enum UnaryOp {
    Neg,        // -
    Not,        // !
    BitNot,     // ~
    Ref,        // &
    Deref,      // *
}

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Type {
    // 기본 타입
    I8,
    I16,
    I32,
    I64,
    U8,
    U16,
    U32,
    U64,
    F32,
    F64,
    Bool,
    String,
    Char,
    Unit,

    // 복합 타입
    Array(Box<Type>),
    Slice(Box<Type>),
    Map {
        key: Box<Type>,
        value: Box<Type>,
    },

    // NULL 안전성
    Option(Box<Type>),
    Result {
        ok: Box<Type>,
        err: Box<Type>,
    },

    // 사용자 정의
    Named(String),

    // 함수
    Function {
        params: Vec<Type>,
        return_type: Box<Type>,
    },

    // 포인터/참조
    Ref(Box<Type>),
    Ptr(Box<Type>),

    // 제네릭
    Generic(String),

    // 기타
    Tuple(Vec<Type>),
    Never,
    Unknown,
}

impl Type {
    /// FreeLang 호환 문자열 표현
    pub fn to_freelang_string(&self) -> String {
        match self {
            Type::I8 => "i8".to_string(),
            Type::I16 => "i16".to_string(),
            Type::I32 => "i32".to_string(),
            Type::I64 => "i64".to_string(),
            Type::U8 => "u8".to_string(),
            Type::U16 => "u16".to_string(),
            Type::U32 => "u32".to_string(),
            Type::U64 => "u64".to_string(),
            Type::F32 => "f32".to_string(),
            Type::F64 => "f64".to_string(),
            Type::Bool => "bool".to_string(),
            Type::String => "String".to_string(),
            Type::Char => "char".to_string(),
            Type::Unit => "()".to_string(),
            Type::Array(inner) => format!("Vec({})", inner.to_freelang_string()),
            Type::Slice(inner) => format!("[{}]", inner.to_freelang_string()),
            Type::Map { key, value } => {
                format!("HashMap({}, {})", key.to_freelang_string(), value.to_freelang_string())
            }
            Type::Option(inner) => format!("Option({})", inner.to_freelang_string()),
            Type::Result { ok, err } => {
                format!("Result({}, {})", ok.to_freelang_string(), err.to_freelang_string())
            }
            Type::Named(name) => name.clone(),
            Type::Function { params, return_type } => {
                let param_str = params.iter()
                    .map(|p| p.to_freelang_string())
                    .collect::<Vec<_>>()
                    .join(", ");
                format!("fn({}) -> {}", param_str, return_type.to_freelang_string())
            }
            Type::Ref(inner) => format!("&{}", inner.to_freelang_string()),
            Type::Ptr(inner) => format!("*{}", inner.to_freelang_string()),
            Type::Generic(name) => format!("T_{}", name),
            Type::Tuple(types) => {
                let type_str = types.iter()
                    .map(|t| t.to_freelang_string())
                    .collect::<Vec<_>>()
                    .join(", ");
                format!("({})", type_str)
            }
            Type::Never => "!".to_string(),
            Type::Unknown => "Unknown".to_string(),
        }
    }
}

// V 타입 → FreeLang 타입 매핑
pub fn v_type_to_freelang(v_type: &str) -> Type {
    match v_type {
        "int" | "i64" => Type::I64,
        "i32" => Type::I32,
        "i16" => Type::I16,
        "i8" => Type::I8,
        "uint" | "u64" => Type::U64,
        "u32" => Type::U32,
        "u16" => Type::U16,
        "u8" => Type::U8,
        "f64" | "f32" => Type::F64,
        "bool" => Type::Bool,
        "string" => Type::String,
        "rune" | "char" => Type::Char,
        name => Type::Named(name.to_string()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_type_conversion() {
        assert_eq!(v_type_to_freelang("int"), Type::I64);
        assert_eq!(v_type_to_freelang("bool"), Type::Bool);
        assert_eq!(v_type_to_freelang("string"), Type::String);
    }

    #[test]
    fn test_freelang_string() {
        assert_eq!(Type::I64.to_freelang_string(), "i64");
        assert_eq!(Type::String.to_freelang_string(), "String");
        assert_eq!(
            Type::Option(Box::new(Type::I64)).to_freelang_string(),
            "Option(i64)"
        );
    }
}
