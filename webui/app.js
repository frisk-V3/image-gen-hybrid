const uploadBtn = document.getElementById("upload");
const filesInput = document.getElementById("files");
const status = document.getElementById("status");
const outputs = document.getElementById("outputs");

let evtSource = new EventSource("/events");
evtSource.onmessage = (e) => {
  try {
    const obj = JSON.parse(e.data);
    if (obj.event === "preprocess_done") {
      status.innerText = `Preprocess done: ${obj.count} images`;
    } else if (obj.event === "generate_output") {
      const payload = JSON.parse(obj.payload);
      if (payload.status === "ok") {
        status.innerText = "Generation finished";
        showOutputs(payload.outputs);
      }
    } else if (obj.event === "error") {
      status.innerText = "Error: " + obj.msg;
    } else if (obj.event === "done") {
      status.innerText = "All done";
    }
  } catch (err) {
    console.error(err);
  }
};

function showOutputs(list) {
  outputs.innerHTML = "";
  list.forEach(p => {
    const img = document.createElement("img");
    img.src = "/static_proxy?path=" + encodeURIComponent(p);
    // static proxy will be implemented by server if needed; for now use file:// fallback
    img.onerror = () => { img.src = p; };
    outputs.appendChild(img);
  });
}

uploadBtn.onclick = async () => {
  const files = filesInput.files;
  if (!files || files.length === 0) {
    status.innerText = "Select files first";
    return;
  }
  const form = new FormData();
  for (let f of files) form.append("files", f, f.name);
  status.innerText = "Uploading...";
  const res = await fetch("/upload", { method: "POST", body: form });
  const j = await res.json();
  if (j.status === "accepted") {
    status.innerText = `Accepted ${j.count} images, generating...`;
  } else {
    status.innerText = "Upload failed";
  }
};
