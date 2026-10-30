//! Star's Rust launcher.
//!
//! This crate is the `star` binary: it starts the Go core without writing to
//! the terminal and performs in-place self-updates.
//!
//! The pieces live in a library so they can be unit tested without spawning a
//! terminal.

pub mod coord;
pub mod handshake;
pub mod platform;
pub mod splash;
pub mod update;
