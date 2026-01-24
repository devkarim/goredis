# Goredis

## Overview
![Redis system](https://i.ibb.co/LDmpYLyK/image.png)

Goredis is a clone of [Redis](https://redis.io) - an In-Memory database.

This project is made for educational purposes only, to grasp how low-level projects work behind the scenes.

Inspired by:
- [Build Redis from scratch by Ahmed Ashraf](https://www.build-redis-from-scratch.dev/)
- [Build Your Own X](https://github.com/codecrafters-io/build-your-own-x)

## Features
- [x] TCP server
- [x] RESP parser
- [x] Command handlers
- [x] Caching layer
- [x] Sharding
- [x] Data persistence layer
- [ ] Set key expiration
- [ ] Pub/Sub
- [ ] Eviction policies
    - [x] FIFO
    - [ ] LRU
- [ ] Configuration file
    - [ ] Port
    - [ ] Memory limit
    - [ ] Verbose flag
- [ ] Sorted sets (Optional)
- [ ] Transactions (Optional)

## Prerequisites
- go v1.24.4
- redis-cli

## How to Run
Simply type the command below inside the project directory to run the server (make sure port 6379 is available):
```
make
```

Then you can directly communicate with the server by using [redis-cli](https://redis.io/docs/latest/develop/tools/cli/) by using the command below:
```
redis-cli -p 6379
```

## References
- [Redis serialization protocol specification](https://redis.io/docs/latest/develop/reference/protocol-spec/#nulls)
- [Concurrency Control in Go: Mastering Mutex and RWMutex for Critical Sections](https://leapcell.io/blog/concurrency-control-in-go-mastering-mutex-and-rwmutex-for-critical-sections)

## License
This project is licensed under MIT license.
