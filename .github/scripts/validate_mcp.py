"""Structural checks of the mcp/ tree, drafts included.

Usage: validate_mcp.py [root]   (default: mcp)

Every file is YAML; every document is an MCPTemplate or a CredentialTemplate
with a metadata.name, and all share one apiVersion, whose last segment is the
tree's release channel. Each file holds exactly one MCPTemplate, named like the
file. The schema itself is checked by the gateway that seeds these documents.
"""
import os
import sys

import yaml

KINDS = {"MCPTemplate", "CredentialTemplate"}
# The gateway ships these and refuses a document of the same name.
BUILTIN_CREDENTIAL_TEMPLATES = {"APIKey", "BasicAuth", "DatabaseDSN", "OAuthApp"}


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "mcp"
    errors = []
    api_version, first = None, None
    seen = {}
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames.sort()
        for name in sorted(filenames):
            path = os.path.join(dirpath, name)
            stem, ext = os.path.splitext(name)
            if ext != ".yaml":
                errors.append(f"{path}: only .yaml files belong under {root}/")
                continue
            try:
                with open(path) as f:
                    docs = [d for d in yaml.safe_load_all(f) if d is not None]
            except yaml.YAMLError as e:
                errors.append(f"{path}: {e}")
                continue
            templates = 0
            for i, doc in enumerate(docs):
                where = f"{path} doc {i}"
                if not isinstance(doc, dict):
                    errors.append(f"{where}: not a mapping")
                    continue
                av, kind = doc.get("apiVersion"), doc.get("kind")
                if not isinstance(av, str) or "/" not in av:
                    errors.append(f"{where}: apiVersion {av!r} is not <group>/<version>")
                elif api_version is None:
                    api_version, first = av, where
                elif av != api_version:
                    errors.append(f"{where}: apiVersion {av!r}, but {first} has {api_version!r}")
                if kind not in KINDS:
                    errors.append(f"{where}: kind {kind!r}, want one of {sorted(KINDS)}")
                    continue
                meta = doc.get("metadata")
                doc_name = meta.get("name") if isinstance(meta, dict) else None
                if not isinstance(doc_name, str) or not doc_name:
                    errors.append(f"{where}: no metadata.name")
                    continue
                if kind == "MCPTemplate":
                    templates += 1
                    if doc_name != stem:
                        errors.append(f"{where}: MCPTemplate {doc_name!r} lives in {name}; name the file {doc_name}.yaml")
                elif doc_name in BUILTIN_CREDENTIAL_TEMPLATES:
                    errors.append(f"{where}: {doc_name} is a built-in CredentialTemplate")
                if (kind, doc_name) in seen:
                    errors.append(f"{where}: {kind} {doc_name!r} is also declared in {seen[(kind, doc_name)]}")
                seen[(kind, doc_name)] = path
            if templates != 1:
                errors.append(f"{path}: holds {templates} MCPTemplates, want exactly one")
    for e in errors:
        print(f"::error::{e}")
    if errors:
        sys.exit(1)
    print(f"{root}/: {len(seen)} documents, apiVersion {api_version}")


if __name__ == "__main__":
    main()
