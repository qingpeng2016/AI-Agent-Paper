#!/usr/bin/env python3
"""[已废弃] paper_llm_model_config 现用明文 api_key 列，无需 seal。

保留脚本仅供历史密文迁移时解密对照；新库请直接在 SQL 写 api_key。
"""

from __future__ import annotations

import hashlib
import secrets
import sys

from cryptography.hazmat.primitives.ciphers.aead import AESGCM

# 等同 conf/local.conf.yaml + NewVaultFromConfig
VAULT_MATERIAL = "ai-agent-paper-api-keys|ai-agent-paper|local|ai_agent_paper"
NONCE_SIZE = 12


def seal(plaintext: str) -> str:
    key = hashlib.sha256(VAULT_MATERIAL.encode("utf-8")).digest()
    aesgcm = AESGCM(key)
    nonce = secrets.token_bytes(NONCE_SIZE)
    body = aesgcm.encrypt(nonce, plaintext.encode("utf-8"), None)
    return (nonce + body).hex()


def main() -> int:
    if len(sys.argv) >= 2:
        plain = sys.argv[1]
    elif not sys.stdin.isatty():
        plain = sys.stdin.read()
    else:
        print("usage: seal_llm_api_key.py <api-key>", file=sys.stderr)
        return 1
    plain = plain.strip()
    if not plain:
        print("error: empty api key", file=sys.stderr)
        return 1
    print(seal(plain))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
