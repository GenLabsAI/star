use std::io::IsTerminal;

pub fn is_tty() -> bool {
    std::io::stdout().is_terminal()
}

#[cfg(target_os = "windows")]
mod imp {
    #[link(name = "kernel32")]
    unsafe extern "system" {
        fn GetStdHandle(n_std_handle: u32) -> *mut std::ffi::c_void;
        fn GetConsoleScreenBufferInfo(
            h_console: *mut std::ffi::c_void,
            lp_info: *mut ConsoleScreenBufferInfo,
        ) -> i32;
    }

    #[repr(C)]
    struct Coord {
        x: i16,
        y: i16,
    }

    #[repr(C)]
    struct SmallRect {
        left: i16,
        top: i16,
        right: i16,
        bottom: i16,
    }

    #[repr(C)]
    struct ConsoleScreenBufferInfo {
        dw_size: Coord,
        dw_cursor_position: Coord,
        w_attributes: u16,
        sr_window: SmallRect,
        dw_maximum_window_size: Coord,
    }

    const STD_OUTPUT_HANDLE: u32 = 0xFFFFFFF5;

    pub fn console_size() -> (u16, u16) {
        let handle = unsafe { GetStdHandle(STD_OUTPUT_HANDLE) };
        let mut info = ConsoleScreenBufferInfo {
            dw_size: Coord { x: 0, y: 0 },
            dw_cursor_position: Coord { x: 0, y: 0 },
            w_attributes: 0,
            sr_window: SmallRect {
                left: 0,
                top: 0,
                right: 0,
                bottom: 0,
            },
            dw_maximum_window_size: Coord { x: 0, y: 0 },
        };
        if unsafe { GetConsoleScreenBufferInfo(handle, &mut info) } == 0 {
            return (80, 24);
        }
        let width = (info.sr_window.right - info.sr_window.left + 1).max(1) as u16;
        let height = (info.sr_window.bottom - info.sr_window.top + 1).max(1) as u16;
        (width, height)
    }
}

#[cfg(unix)]
mod imp {
    #[repr(C)]
    struct WindowSize {
        rows: u16,
        columns: u16,
        x_pixels: u16,
        y_pixels: u16,
    }

    unsafe extern "C" {
        fn ioctl(fd: i32, request: usize, ...) -> i32;
    }

    #[cfg(target_os = "linux")]
    const TIOCGWINSZ: usize = 0x5413;
    #[cfg(target_os = "macos")]
    const TIOCGWINSZ: usize = 0x40087468;

    pub fn console_size() -> (u16, u16) {
        let mut size = WindowSize {
            rows: 0,
            columns: 0,
            x_pixels: 0,
            y_pixels: 0,
        };
        if unsafe { ioctl(1, TIOCGWINSZ, &mut size) } == 0 && size.columns > 0 && size.rows > 0 {
            return (size.columns, size.rows);
        }
        (80, 24)
    }
}

pub use imp::console_size;
