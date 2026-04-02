import tempfile
import os
from model_stub import generate_batch

def test_generate_batch_creates_files(tmp_path):
    # create dummy thumb files
    thumbs = []
    for i in range(2):
        p = tmp_path / f"t{i}.png"
        p.write_bytes(b"\x89PNG\r\n\x1a\n")
        thumbs.append({"thumb_path": str(p)})
    out = generate_batch(["p1","p2"], thumbs)
    assert len(out) == 2
    for p in out:
        assert os.path.exists(p)
