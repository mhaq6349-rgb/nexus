use crate::types::{self, Value};
use std::collections::HashMap;
use std::sync::Mutex;
use std::time::Instant;

type FunctionHandler = Box<dyn Send + Fn(Vec<Value>) -> Result<Value, String>>;

static REGISTRY: once_cell::sync::Lazy<Mutex<HashMap<String, FunctionHandler>>> =
    once_cell::sync::Lazy::new(|| Mutex::new(HashMap::new()));

pub fn register(name: &str, handler: FunctionHandler) {
    REGISTRY.lock().unwrap().insert(name.to_string(), handler);
}

pub fn call_function(name: &str, args: Vec<Value>) -> Result<Value, String> {
    let registry = REGISTRY.lock().unwrap();
    match registry.get(name) {
        Some(handler) => handler(args),
        None => Err(format!("function '{}' not registered", name)),
    }
}

pub fn init_default_functions() {
    register("math.add", Box::new(|args| {
        let sum: f64 = args.iter().filter_map(|a| a.as_f64()).sum();
        Ok(Value::F64(sum))
    }));
    register("math.mul", Box::new(|args| {
        let prod: f64 = args.iter().filter_map(|a| a.as_f64()).product();
        Ok(Value::F64(prod))
    }));
    register("math.simd_mul", Box::new(|args| {
        if args.len() < 2 { return Err("need array + scalar".into()); }
        let arr = match &args[0] { Value::NdArray { data, .. } => data.clone(), _ => return Err("first arg not array".into()) };
        let scalar = args[1].as_f64().unwrap_or(1.0) as f32;
        let n = arr.len() / 4;
        let mut inp = vec![0f32; n];
        let mut out = vec![0f32; n];
        for (i, chunk) in arr.chunks_exact(4).enumerate().take(n) {
            inp[i] = f32::from_le_bytes([chunk[0], chunk[1], chunk[2], chunk[3]]);
        }
        unsafe { crate::ffi::nexus_apply_simd_f32(inp.as_ptr(), out.as_mut_ptr(), n as u32, 1); }
        let result_bytes: Vec<u8> = out.iter().flat_map(|v| v.to_le_bytes()).collect();
        Ok(Value::NdArray { dtype: 1, shape: vec![n as u64], data: result_bytes })
    }));
    register("string.reverse", Box::new(|args| {
        let s = args.first().and_then(|a| a.as_str()).unwrap_or("");
        Ok(Value::String(s.chars().rev().collect()))
    }));
    register("string.concat", Box::new(|args| {
        let result: String = args.iter().filter_map(|a| a.as_str()).collect();
        Ok(Value::String(result))
    }));
    register("system.echo", Box::new(|args| Ok(args.into_iter().next().unwrap_or(Value::Null))));
    register("system.ping", Box::new(|_| Ok(Value::String("pong".into()))));
}

pub fn process_message(type_id: u32, data: &[u8]) -> Option<Vec<u8>> {
    match type_id {
        types::MSG_CALL => {
            if let Some(msg) = types::deserialize_call(data) {
                let start = Instant::now();
                let result = call_function(&msg.function, msg.args);
                let elapsed = start.elapsed().as_micros();
                match result {
                    Ok(val) => {
                        let mut resp = types::serialize_return(msg.call_id, val);
                        resp.extend_from_slice(&elapsed.to_le_bytes());
                        Some(resp)
                    }
                    Err(e) => Some(types::serialize_error(msg.call_id, -1, &e)),
                }
            } else { None }
        }
        types::MSG_PING => Some(types::serialize_return(0, Value::String("pong".into()))),
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_registry() {
        init_default_functions();
        let r = call_function("math.add", vec![Value::F64(1.0), Value::F64(2.0)]);
        assert_eq!(r.unwrap().as_f64(), Some(3.0));
        let r = call_function("string.reverse", vec![Value::String("hello".into())]);
        assert_eq!(r.unwrap().as_str(), Some("olleh"));
        let r = call_function("system.ping", vec![]);
        assert_eq!(r.unwrap().as_str(), Some("pong"));
    }
}
