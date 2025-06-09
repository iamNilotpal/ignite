# ignite - A High-Throughput In-Memory Key-Value Store

`ignite` is a high-performance key/value data store designed for fast read and
write operations inspired by [Bitcask](https://en.wikipedia.org/wiki/Bitcask).
It combines an in-memory hash table (KeyDir) with an append-only log structure
on disk to achieve high throughput. It is designed for applications requiring
fast read and write operations, such as caching, session management, and
real-time data processing. The goal is to provide a simple, efficient, and
reliable solution for in-memory data storage in Go applications.

---

## Design Goals

1. **High Write Throughput**: Achieved by sequential, append-only writes to
   disk.
2. **Fast Reads**: Enabled by an in-memory hash table (KeyDir) that points
   directly to the data location on disk.
3. **Durability**: Data is immediately written to disk, ensuring durability.
4. **Fast Startup**: Hint files store metadata to rebuild the KeyDir quickly
   after a restart.
5. **Simplicity**: Avoid complex data structures or algorithms to keep the
   system maintainable and predictable.

---

## Architecture

Ignite's architecture consists of two main parts:

1. **On-Disk Storage**: An append-only log where data is written sequentially.
2. **In-Memory Index (KeyDir)**: A hash table that maps keys to their locations
   on disk.

```
┌──────────────────────┐       ┌──────────────────────┐
│                      │       │                      │
│      In-Memory       │       │       On-Disk        │
│       KeyDir         │       │   Append-Only Log    │
│                      │       │                      │
└──────────┬───────────┘       └──────────┬───────────┘
           │                              │
           │                              │
           └───────────┐      ┌───────────┘
                       │      │
                   ┌───▼──────▼───┐
                   │              │
                   │   Ignite     │
                   │   Engine     │
                   │              │
                   └──────────────┘
```

---

## Key Components

### Storage Format

Data is stored on disk in segments (files). At any given time, only one segment
is active for writes. Each entry in the segment has the following structure:

### On-Disk Entry Layout

```
+=======================+=======================+=======================+=====
|       Entry 1         |       Entry 2         |       Entry 3         | ...
+=======================+=======================+=======================+=====
  |
  +-> Structure of Entry 1:
      +----------+-----------+---------+---------+-----------+-----+-------+
      | Checksum | Timestamp | Version | KeySize | ValueSize | Key | Value |
      +----------+-----------+---------+---------+-----------+-----+-------+
      |<--------------------- Header ----------------------->|<- Payload ->|
```

- **Checksum**: Ensures data integrity.
- **Timestamp**: When the entry was written.
- **Version**: Version of the entry format (for backward compatibility).
- **KeySize/ValueSize**: Lengths of the key and value.
- **Key/Data**: The actual key and value bytes.

---

### KeyDir (In-Memory Hash Table)

The KeyDir is an in-memory hash table that maps keys to their locations on disk.
It is rebuilt during startup using hint files (explained below).

### KeyDir Lookup Flow

1. A key is provided for lookup.
2. KeyDir is queried to find the `SegmentId` and `Offset`.
3. The segment file is read at the specified offset to retrieve the value.

---

### Hint Files

Hint files are an optimization to speed up startup. They store the metadata
needed to rebuild the KeyDir without scanning all data files.

### Hint File Entry Layout

```
┌───────────┬───────────┬───────────┬───────────┬───────────┐
│ KeySize   │ Key       │ SegmentId │ Offset    │ Timestamp │
│ (4 bytes) │ (N bytes) │ (2 bytes) │ (8 bytes) │ (8 bytes) │
└───────────┴───────────┴───────────┴───────────┴───────────┘
```

- **KeySize/Key**: The key and its length.
- **SegmentId/Offset**: Location of the entry in the data file.
- **Timestamp**: When the entry was written.

---

## Write and Read Paths

### Write Path

1. The key/value is serialized into the on-disk format.
2. The entry is appended to the active segment file.
3. The KeyDir is updated with the new entry's location.

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│             │    │             │    │             │
│  Serialize  │───▶│ Append to   │───▶│ Update      │
│  Entry      │    │ Segment     │    │ KeyDir      │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Read Path

1. The KeyDir is queried for the key.
2. If found, the segment file is read at the specified offset.
3. The value is returned to the caller.

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│             │    │             │    │             │
│  Query      │───▶│ Read from   │───▶│ Return      │
│  KeyDir     │    │ Segment     │    │ Value       │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
```

---

## Compaction

Over time, segments accumulate stale data (e.g., overwritten or deleted keys).
Compaction merges segments, retaining only the latest versions of keys.

1. **Select Segments**: Identify segments with the most stale data.
2. **Merge**: Create a new segment with only the latest key versions.
3. **Update KeyDir**: Point keys to the new segment.
4. **Delete Old Segments**: Remove the old segments.

```
┌───────────────────────┐       ┌───────────────────────┐
│                       │       │                       │
│  Segment 1 (Old)      │       │  Segment 2 (Old)      │
│  - KeyA: "value1"     │       │  - KeyA: "value2"     │
│  - KeyB: "valueX"     │       │  - KeyC: "valueY"     │
│                       │       │                       │
└───────────┬───────────┘       └───────────┬───────────┘
            │                                │
            └───────────────┐    ┌───────────┘
                            │    │
                        ┌──▼────▼───┐
                        │           │
                        │  Merged   │
                        │  Segment  │
                        │  - KeyA: "value2" │
                        │  - KeyB: "valueX" │
                        │  - KeyC: "valueY" │
                        │           │
                        └───────────┘
```

---

## Performance Trade-offs

1. **Write Performance**: Append-only writes are fast but require compaction to
   reclaim space.
2. **Read Performance**: KeyDir enables O(1) lookups but consumes memory.
3. **Startup Time**: Hint files reduce startup time but add complexity.
4. **Disk Space**: Compaction reduces disk usage but is CPU-intensive.

---

## Diagrams

### Data Flow Diagram

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│             │    │             │    │             │
│   Client    │───▶│   Ignite    │───▶│   Disk      │
│  (Read/Write)│    │  (KeyDir)   │    │ (Segments)  │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Segment File Layout

```
┌───────────────────────────────────────────────────────┐
│ Entry 1 (Header + Key + Data)                         │
├───────────────────────────────────────────────────────┤
│ Entry 2 (Header + Key + Data)                         │
├───────────────────────────────────────────────────────┤
│ ...                                                   │
└───────────────────────────────────────────────────────┘
```

---

## Conclusion

Ignite's design prioritizes simplicity and performance. By combining an
in-memory hash table with append-only disk writes, it achieves high throughput
for both reads and writes. Compaction and hint files address long-term storage
and startup efficiency. This blueprint provides a foundation for building a
robust key/value store in Go.
