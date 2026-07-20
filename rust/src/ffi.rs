use crate::memory::SharedMemory;
use crate::types;
use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::sync::Mutex;

static MEM: once_cell::sync::Lazy<Mutex<SharedMemory>> =
    once_cell::sync::Lazy::new(|| Mutex::new(SharedMemory::new()));

#[no_mangle]
pub extern "C" fn nexus_init() -> i32 {
    let _ = MEM.lock();
    0
}

#[no_mangle]
pub extern "C" fn nexus_write_call(
    call_id: u64,
    function: *const c_char,
    args_data: *const u8,
    args_len: u32,
    source_lang: u8,
) -> i32 {
    let func = if function.is_null() { "" } else {
        unsafe { match CStr::from_ptr(function).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        }}
    };
    let args: Vec<types::Value> = if args_data.is_null() || args_len == 0 {
        vec![]
    } else {
        let slice = unsafe { std::slice::from_raw_parts(args_data, args_len as usize) };
        bincode::deserialize(slice).unwrap_or_default()
    };
    let data = types::serialize_call(call_id, func, args);
    let mut mem = MEM.lock().unwrap();
    if mem.write_slot(types::MSG_CALL, source_lang, &data).is_some() { 0 } else { -2 }
}

#[no_mangle]
pub extern "C" fn nexus_write_return(call_id: u64, result_data: *const u8, result_len: u32) -> i32 {
    let value: types::Value = if result_data.is_null() || result_len == 0 {
        types::Value::Null
    } else {
        let slice = unsafe { std::slice::from_raw_parts(result_data, result_len as usize) };
        bincode::deserialize(slice).unwrap_or(types::Value::Null)
    };
    let data = types::serialize_return(call_id, value);
    let mut mem = MEM.lock().unwrap();
    if mem.write_slot(types::MSG_RETURN, types::LANG_RUST, &data).is_some() { 0 } else { -2 }
}

#[no_mangle]
pub extern "C" fn nexus_read_message(
    type_id: *mut u32,
    source_lang: *mut u8,
    buf: *mut u8,
    buf_len: *mut u32,
) -> i32 {
    let mut mem = MEM.lock().unwrap();
    match mem.read_slot() {
        Some((tid, lang, data)) => {
            unsafe {
                *type_id = tid;
                *source_lang = lang;
                let copy_len = data.len().min(*buf_len as usize);
                std::ptr::copy_nonoverlapping(data.as_ptr(), buf, copy_len);
                *buf_len = copy_len as u32;
            }
            0
        }
        None => 1,
    }
}

#[no_mangle]
pub extern "C" fn nexus_apply_simd_f32(
    input: *const f32,
    output: *mut f32,
    len: u32,
    op: u8,
) -> i32 {
    let n = len as usize;
    let inp = unsafe { std::slice::from_raw_parts(input, n) };
    let out = unsafe { std::slice::from_raw_parts_mut(output, n) };
    match op {
        0 => out.copy_from_slice(inp),
        1 => multiply_scalar(inp, out, 2.0),
        2 => relu(inp, out),
        3 => sigmoid_approx(inp, out),
        _ => return -1,
    }
    0
}

fn multiply_scalar(inp: &[f32], out: &mut [f32], scalar: f32) {
    #[cfg(any(target_arch = "x86_64", target_arch = "aarch64"))]
    {
        use wide::f32x4;
        let s = f32x4::splat(scalar);
        for (chunk_in, chunk_out) in inp.chunks_exact(4).zip(out.chunks_exact_mut(4)) {
            let a = f32x4::from(chunk_in);
            (a * s).store(chunk_out);
        }
    }
    #[cfg(not(any(target_arch = "x86_64", target_arch = "aarch64")))]
    {
        for (i, v) in out.iter_mut().enumerate() {
            *v = inp[i] * scalar;
        }
    }
    let rem = inp.len() % 4;
    let start = inp.len() - rem;
    for i in 0..rem {
        out[start + i] = inp[start + i] * scalar;
    }
}

fn relu(inp: &[f32], out: &mut [f32]) {
    #[cfg(any(target_arch = "x86_64", target_arch = "aarch64"))]
    {
        use wide::f32x4;
        let zero = f32x4::splat(0.0);
        for (chunk_in, chunk_out) in inp.chunks_exact(4).zip(out.chunks_exact_mut(4)) {
            let a = f32x4::from(chunk_in);
            a.max(zero).store(chunk_out);
        }
    }
    #[cfg(not(any(target_arch = "x86_64", target_arch = "aarch64")))]
    {
        for (i, v) in out.iter_mut().enumerate() {
            *v = inp[i].max(0.0);
        }
    }
    let rem = inp.len() % 4;
    let start = inp.len() - rem;
    for i in 0..rem {
        out[start + i] = inp[start + i].max(0.0);
    }
}

fn sigmoid_approx(inp: &[f32], out: &mut [f32]) {
    for (i, v) in out.iter_mut().enumerate() {
        let x = inp[i];
        *v = 1.0 / (1.0 + (-x).exp());
    }
}
