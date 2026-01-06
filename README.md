# Goredis

## Overview
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
- [ ] Data persistence layer
- [ ] Set key expiration
- [ ] Pub/Sub
- [ ] Eviction policy such as LRU
- [ ] Configuration file
    - [ ] Port
    - [ ] Memory limit
    - [ ] Verbose flag
- [ ] Sorted sets (Optional)
- [ ] Transactions (Optional)

## Prerequisites
- Go v1.24.4

## How to Run
Simply type the command below inside the project directory:
```
make
```

## License
This project is licensed under MIT license.
