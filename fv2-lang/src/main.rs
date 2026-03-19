mod lexer;
mod ast;

use std::env;
use std::fs;
use lexer::{Lexer, Token};
use ast::Program;

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: fv2 <source.fv>");
        std::process::exit(1);
    }

    let filename = &args[1];
    match fs::read_to_string(filename) {
        Ok(source) => {
            match compile(&source) {
                Ok(c_code) => {
                    println!("{}", c_code);
                }
                Err(e) => {
                    eprintln!("Compilation error: {}", e);
                    std::process::exit(1);
                }
            }
        }
        Err(e) => {
            eprintln!("Error reading file '{}': {}", filename, e);
            std::process::exit(1);
        }
    }
}

fn compile(source: &str) -> Result<String, String> {
    // Step 1: Lexing
    let lexer = Lexer::new(source)?;
    let tokens = lexer.tokenize()?;

    println!("// ===== V-Compatible Lexer =====");
    println!("// Tokenized {} tokens", tokens.len());
    for (i, token) in tokens.iter().take(20).enumerate() {
        println!("// Token {}: {:?}", i, token.token);
    }

    // Step 2: Parser (TODO)
    // Step 3: Type Checker (TODO)
    // Step 4: Code Generator (TODO)

    Ok("// FV 2.0 Compiler - Phase 2 in progress\n// C code generation will be implemented here".to_string())
}
