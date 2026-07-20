use std::alloc::{alloc, dealloc, Layout};
use std::ptr::NonNull;
use std::sync::atomic::{AtomicU64, Ordering};

pub const MAGIC: u32 = 0x4E455855;
pub const VERSION: u32 = 1;
pub const RING_SLOTS: u32 = 256;
pub const SLOT_SIZE: u32 = 512;
pub const ARENA_SIZE: u64 = 16 * 1024 * 1024;

#[repr(C)]
pub struct SharedHeader {
    pub magic: u32,
    pub version: u32,
    pub num_slots: u32,
    pub slot_size: u32,
    pub arena_offset: u64,
    pub arena_size: u64,
    pub head: AtomicU64,
    pub tail: AtomicU64,
    _pad: [u8; 32],
}

#[repr(C)]
pub struct SlotHeader {
    pub len: u32,
    pub type_id: u32,
    pub source_lang: u8,
    pub flags: u8,
    pub _pad: [u8; 2],
}

pub struct SharedMemory {
    ptr: NonNull<u8>,
    layout: Layout,
    size: usize,
}

unsafe impl Send for SharedMemory {}
unsafe impl Sync for SharedMemory {}

impl SharedMemory {
    pub fn new() -> Self {
        let total_size = std::mem::size_of::<SharedHeader>()
            + (RING_SLOTS as usize) * (SLOT_SIZE as usize)
            + ARENA_SIZE as usize;
        let layout = Layout::from_size_align(total_size, 64).unwrap();
        let ptr = unsafe { alloc(layout) };
        let ptr = NonNull::new(ptr).expect("shared memory allocation failed");
        let mut mem = SharedMemory { ptr, layout, size: total_size };
        mem.init_header();
        mem
    }

    fn init_header(&mut self) {
        let header = self.header_mut();
        header.magic = MAGIC;
        header.version = VERSION;
        header.num_slots = RING_SLOTS;
        header.slot_size = SLOT_SIZE;
        header.arena_offset = std::mem::size_of::<SharedHeader>() as u64;
        header.arena_size = ARENA_SIZE;
        header.head.store(0, Ordering::SeqCst);
        header.tail.store(0, Ordering::SeqCst);
    }

    pub fn header(&self) -> &SharedHeader {
        unsafe { &*(self.ptr.as_ptr() as *const SharedHeader) }
    }

    pub fn header_mut(&mut self) -> &mut SharedHeader {
        unsafe { &mut *(self.ptr.as_ptr() as *mut SharedHeader) }
    }

    pub fn slot_offset(&self, index: u32) -> usize {
        std::mem::size_of::<SharedHeader>() + (index as usize) * (SLOT_SIZE as usize)
    }

    pub fn slot_header(&self, index: u32) -> &SlotHeader {
        let off = self.slot_offset(index);
        unsafe { &*(self.ptr.as_ptr().add(off) as *const SlotHeader) }
    }

    pub fn slot_header_mut(&mut self, index: u32) -> &mut SlotHeader {
        let off = self.slot_offset(index);
        unsafe { &mut *(self.ptr.as_ptr().add(off) as *mut SlotHeader) }
    }

    pub fn slot_data(&self, index: u32) -> &[u8] {
        let off = self.slot_offset(index) + std::mem::size_of::<SlotHeader>();
        let len = self.slot_header(index).len as usize;
        unsafe { std::slice::from_raw_parts(self.ptr.as_ptr().add(off), len) }
    }

    pub fn slot_data_mut(&mut self, index: u32) -> &mut [u8] {
        let off = self.slot_offset(index) + std::mem::size_of::<SlotHeader>();
        let max = (SLOT_SIZE as usize) - std::mem::size_of::<SlotHeader>();
        unsafe { std::slice::from_raw_parts_mut(self.ptr.as_ptr().add(off), max) }
    }

    pub fn arena_ptr(&self) -> *mut u8 {
        let off = self.header().arena_offset as usize;
        unsafe { self.ptr.as_ptr().add(off) as *mut u8 }
    }

    pub fn write_slot(&mut self, type_id: u32, source_lang: u8, data: &[u8]) -> Option<u32> {
        let header = self.header_mut();
        let head = header.head.load(Ordering::Acquire);
        let tail = header.tail.load(Ordering::Acquire);
        if head.wrapping_sub(tail) as u64 >= RING_SLOTS as u64 {
            return None;
        }
        let idx = (head % RING_SLOTS as u64) as u32;
        let slot_header = self.slot_header_mut(idx);
        let max_data = (SLOT_SIZE as usize) - std::mem::size_of::<SlotHeader>();
        let len = data.len().min(max_data);
        slot_header.len = len as u32;
        slot_header.type_id = type_id;
        slot_header.source_lang = source_lang;
        slot_header.flags = 0;
        let dst = self.slot_data_mut(idx);
        dst[..len].copy_from_slice(&data[..len]);
        header.head.fetch_add(1, Ordering::Release);
        Some(idx)
    }

    pub fn read_slot(&mut self) -> Option<(u32, u8, Vec<u8>)> {
        let header = self.header_mut();
        let tail = header.tail.load(Ordering::Acquire);
        let head = header.head.load(Ordering::Acquire);
        if tail >= head {
            return None;
        }
        let idx = (tail % RING_SLOTS as u64) as u32;
        let slot_header = self.slot_header(idx);
        let type_id = slot_header.type_id;
        let source_lang = slot_header.source_lang;
        let data = self.slot_data(idx).to_vec();
        header.tail.fetch_add(1, Ordering::Release);
        Some((type_id, source_lang, data))
    }
}

impl Drop for SharedMemory {
    fn drop(&mut self) {
        unsafe { dealloc(self.ptr.as_ptr(), self.layout) }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_ring_buffer() {
        let mut mem = SharedMemory::new();
        assert_eq!(mem.header().magic, MAGIC);
        assert!(mem.write_slot(1, 0, b"hello from rust").is_some());
        let (type_id, _, data) = mem.read_slot().unwrap();
        assert_eq!(type_id, 1);
        assert_eq!(&data, b"hello from rust");
    }
}
