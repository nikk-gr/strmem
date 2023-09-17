# Structured in-memory cache

**strmem** is a library that helps you to make high performance structured in-memory cache

It can give you log(n) access to the line of the cache array by any field of this line

---

## How it works
**strmem** works like an in-memory database. 
It has an array of the data and indexes. 
Supported the following indexes:
1. BTree

To be implemented:
1. Hash index
2. RD-tree for text search
3. Fuzzy string search
    