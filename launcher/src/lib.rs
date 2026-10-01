//! Star's Rust launcher.
//!
//! This crate is the `star` binary: it renders the splash banner, supervises
//! the Go core, and performs in-place self-updates.
//!
//! The pieces live in a library so they can be unit tested and rendered
//! offline (see the `splash_filmstrip` example) without spawning a terminal.

pub mod coord;
pub mod handshake;
pub mod platform;
pub mod splash;
pub mod update;
