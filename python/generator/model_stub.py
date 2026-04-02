from typing import List, Dict
import shutil
import os
import uuid

def generate_batch(prompts: List[str], batch_meta: List[Dict]) -> List[str]:
    """
    Stub: pretend to generate images and return output paths.
    Implementation: copy thumb to a generated file path to simulate output.
    """
    out = []
    outdir = os.path.join("/tmp", "image-gen-gui-out")
    os.makedirs(outdir, exist_ok=True)
    for i, meta in enumerate(batch_meta):
        src = meta.get("thumb_path")
        dst = os.path.join(outdir, f"gen_{uuid.uuid4().hex}.png")
        try:
            shutil.copyfile(src, dst)
        except Exception:
            # fallback: create empty file
            open(dst, "wb").close()
        out.append(dst)
    return out
