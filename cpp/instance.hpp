//
// `swallow` - 'trace log storage for recommender system'
// Copyright (C) 2019 - present timepi <timepi123@gmail.com>
// LuBan is provided under: GNU Affero General Public License (AGPL3.0)
// https://www.gnu.org/licenses/agpl-3.0.html unless stated otherwise.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be usefulType,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//

#ifndef SWALLOW_INSTANCE_HPP
#define SWALLOW_INSTANCE_HPP

#pragma once

#include <rocksdb/db.h>
#include <rocksdb/write_batch.h>

#include <iostream>
#include <string>
#include <vector>

namespace swallow {

class Instance;

class Request {
 public:
  Request() = default;
  ~Request() = default;
  void append(rocksdb::Slice key, rocksdb::Slice value) {
    batch_.Put(key, value);
  }

 private:
  rocksdb::WriteBatch batch_;
  friend class Instance;
};

class Instance {
 public:
  Instance() = delete;
  Instance(std::string data_dir) : db_(nullptr) {
    rocksdb::Options options;
    options.create_if_missing = true;
    options.disable_auto_compactions = true;
    rocksdb::Status status = rocksdb::DB::Open(options, data_dir, &db_);
    if (!status.ok()) {
      std::cerr << "open leveldb error: " << status.ToString() << std::endl;
      throw std::runtime_error("open rocksdb error");
    }
    assert(db_ != nullptr);
  }
  ~Instance() {
    db_->Close();
    delete db_;
  }

  void compcat() {
    rocksdb::CompactRangeOptions compact_options;
    compact_options.target_level = -1;
    compact_options.bottommost_level_compaction =
        rocksdb::BottommostLevelCompaction::kSkip;
    db_->CompactRange(compact_options, nullptr, nullptr);
  }

  void put(Request *req) {
    rocksdb::WriteOptions options;
    options.sync = false;
    db_->Write(options, &req->batch_);
  }

 private:
  rocksdb::DB *db_;
};
}  // namespace swallow

#endif  // SWALLOW_INSTANCE_HPP