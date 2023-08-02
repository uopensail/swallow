#include "swallow.h"

#include "instance.hpp"

void *swallow_open(void *dir, unsigned long long len) {
  try {
    return new swallow::Instance({(char *)dir, len});
  } catch (...) {
    return nullptr;
  }
}

void swallow_close(void *ins) {
  if (ins == nullptr) {
    return;
  }
  swallow::Instance *instance = (swallow::Instance *)ins;
  delete instance;
}

void swallow_compact(void *ins) {
  if (ins == nullptr) {
    return;
  }
  swallow::Instance *instance = (swallow::Instance *)ins;
  instance->compcat();
}

void *swallow_put(void *ins, void *req) {
  if (ins == nullptr || req == nullptr) {
    return;
  }
  swallow::Instance *instance = (swallow::Instance *)ins;
  instance->put((swallow::Request *)req);
}

void *swallow_new_request() { return new swallow::Request(); }

void swallow_del_request(void *req) {
  if (req == nullptr) {
    return;
  }
  swallow::Request *r = (swallow::Request *)req;
  delete r;
}

void swallow_request_append(void *req, void *key, unsigned long long klen,
                            void *value, unsigned long vlen) {
  if (req == nullptr) {
    return;
  }
  swallow::Request *r = (swallow::Request *)req;
  r->append({(char *)key, klen}, {(char *)value, vlen});
}