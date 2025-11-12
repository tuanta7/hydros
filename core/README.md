# OAuth Core Package

Top-to-bottom: higher layers may import lower layers; lower layers must not import higher layers.

- **handlers**: Implement the OAuth handlers as defined in the core package.
- **strategy**: Defines strategies for token management, validation, etc.
- **storage**: Defines storage interfaces for interacting with data stores.
- **signer**: Defines opaque and jwt token signers.
- **core**: Defines the core interfaces, types and behaviors of OAuth handlers
- **x**: Shared utilities