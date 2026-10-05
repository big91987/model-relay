# Cross-layer checklist

- Does the change alter HTTP payload, status, method, CSRF, storage or config? If not, keep it in the owning DOM layer.
- If yes, trace CLI/config → Store → Server → UI; identify signatures, validation and failure mapping.
- Can failure leak a credential, prompt or raw upstream diagnostic?
- Which handler and actual HTTP regression tests enforce the contract?
- Follow ../backend/runtime-contracts.md; do not silently broaden a frontend task into authentication changes.
