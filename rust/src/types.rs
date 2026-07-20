use serde::{Deserialize, Serialize};

pub const LANG_RUST: u8 = 0;
pub const LANG_GO: u8 = 1;
pub const LANG_PYTHON: u8 = 2;
pub const LANG_TS: u8 = 3;

pub const MSG_CALL: u32 = 0x0001;
pub const MSG_RETURN: u32 = 0x0002;
pub const MSG_ERROR: u32 = 0x0003;
pub const MSG_PING: u32 = 0x00FE;
pub const MSG_PONG: u32 = 0x00FF;
pub const MSG_LOG: u32 = 0x0100;
pub const MSG_METRIC: u32 = 0x0101;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CallMessage {
    pub call_id: u64,
    pub function: String,
    pub args: Vec<Value>,
    pub timeout_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReturnMessage {
    pub call_id: u64,
    pub result: Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorMessage {
    pub call_id: u64,
    pub code: i32,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogMessage {
    pub level: u8,
    pub source: String,
    pub text: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricMessage {
    pub name: String,
    pub value: f64,
    pub labels: Vec<(String, String)>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Value {
    Null,
    Bool(bool),
    I64(i64),
    U64(u64),
    F64(f64),
    String(String),
    Bytes(Vec<u8>),
    List(Vec<Value>),
    Map(Vec<(String, Value)>),
    NdArray { dtype: u8, shape: Vec<u64>, data: Vec<u8> },
}

impl Value {
    pub fn as_i64(&self) -> Option<i64> {
        if let Value::I64(v) = self { Some(*v) } else { None }
    }
    pub fn as_f64(&self) -> Option<f64> {
        if let Value::F64(v) = self { Some(*v) } else { None }
    }
    pub fn as_str(&self) -> Option<&str> {
        if let Value::String(v) = self { Some(v) } else { None }
    }
    pub fn as_bytes(&self) -> Option<&[u8]> {
        if let Value::Bytes(v) = self { Some(v) } else { None }
    }
}

pub fn serialize_call(id: u64, func: &str, args: Vec<Value>) -> Vec<u8> {
    let msg = CallMessage { call_id: id, function: func.to_string(), args, timeout_ms: 5000 };
    bincode::serialize(&msg).unwrap_or_default()
}

pub fn serialize_return(id: u64, result: Value) -> Vec<u8> {
    let msg = ReturnMessage { call_id: id, result };
    bincode::serialize(&msg).unwrap_or_default()
}

pub fn serialize_error(id: u64, code: i32, msg: &str) -> Vec<u8> {
    let msg = ErrorMessage { call_id: id, code, message: msg.to_string() };
    bincode::serialize(&msg).unwrap_or_default()
}

pub fn deserialize_call(data: &[u8]) -> Option<CallMessage> {
    bincode::deserialize(data).ok()
}

pub fn deserialize_return(data: &[u8]) -> Option<ReturnMessage> {
    bincode::deserialize(data).ok()
}

pub fn deserialize_error(data: &[u8]) -> Option<ErrorMessage> {
    bincode::deserialize(data).ok()
}
