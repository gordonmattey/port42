use std::env;
use std::fs;

fn main() {
    // Read version from version.txt in the repository root
    let version = if let Ok(contents) = fs::read_to_string("../version.txt") {
        contents.trim().to_string()
    } else {
        // Fallback to Cargo.toml version
        env!("CARGO_PKG_VERSION").to_string()
    };
    
    // Set VERSION environment variable for compile time
    println!("cargo:rustc-env=PORT42_VERSION={}", version);
    
    // Rerun if version.txt changes
    println!("cargo:rerun-if-changed=../version.txt");
}