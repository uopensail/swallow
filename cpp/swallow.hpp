//
// `swallow` - 'trace server for recommender system'
// Copyright (C) 2019 - present timepi <timepi123@gmail.com>
// `Damo-Embedding` is provided under: GNU Affero General Public License
// (AGPL3.0) https://www.gnu.org/licenses/agpl-3.0.html unless stated otherwise.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//

#ifndef SWALLOW_INSTANCE_HPP
#define SWALLOW_INSTANCE_HPP

#pragma once

#include <iostream>
#include <rocksdb/db.h>
#include <rocksdb/merge_operator.h>
#include <rocksdb/utilities/db_ttl.h>
#include <rocksdb/write_batch.h>
#include <string>
#include <string_view>
#include <vector>

const size_t max_count = 1000;

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

class Response {
public:
  Response() = default;
  ~Response() = default;
  const std::string &operator[](size_t i) const { return data_[i]; }

private:
  std::vector<std::string> data_;
  friend class Instance;
};

class Instance {
public:
  Instance() = delete;
  Instance(std::string data_dir, bool writable)
      : db_(nullptr), writable_(writable) {
    rocksdb::Status status;
    if (writable) {
      rocksdb::Options options;
      options.create_if_missing = true;
      options.disable_auto_compactions = true;
      status = rocksdb::DB::Open(options, data_dir, &db_);
    } else {
      status = rocksdb::DB::OpenForReadOnly(rocksdb::Options(), data_dir, &db_);
    }

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
    if (writable_) {
      db_->CompactRange(rocksdb::CompactRangeOptions(), nullptr, nullptr);
    }
  }

  Response *scan(rocksdb::Slice start, rocksdb::Slice end) {
    Response *resp = new Response();
    const rocksdb::Snapshot *sp = db_->GetSnapshot();
    rocksdb::ReadOptions options;
    // include start
    options.iterate_lower_bound = &start;
    // exclude end
    options.iterate_upper_bound = &end;
    options.snapshot = sp;
    rocksdb::Iterator *it = this->db_->NewIterator(options);
    it->SeekToFirst();

    if (it->Valid()) {
      if (it->value() == start) {
        it->Next();
      }
    }
    size_t count = 0;
    for (; it->Valid() && count < max_count; it->Next()) {
      resp->data_.emplace_back(it->value());
      count++;
    }
    assert(it->status().ok());
    delete it;
    db_->ReleaseSnapshot(sp);
    return resp;
  }

  void put(Request *req) {
    if (writable_) {
      rocksdb::WriteOptions options;
      options.sync = false;
      db_->Write(options, &req->batch_);
    }
  }

private:
  rocksdb::DB *db_;
  bool writable_;
};

#endif // SWALLOW_INSTANCE_HPP
