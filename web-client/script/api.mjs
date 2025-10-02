function timeString(duration) {
  const seconds = duration % 60; duration = (duration - seconds) / 60;
  const minutes = duration % 60; duration = (duration - minutes) / 60;
  const hours   = duration;
  var time_string = `${seconds}s`;
  if(minutes > 0) { time_string = `${minutes}m ${time_string}`; }
  if(hours   > 0) { time_string = `${hours}h ${time_string}`; }
  return time_string;
}

function dateString(unix_seconds) {
  if(unix_seconds == 0) { return ""; }

  const date_time = new Date(unix_seconds * 1000);
  const year   = date_time.getFullYear();
  const month  = (date_time.getMonth() + 1).toString().padStart(2, "0");
  const date   = date_time.getDate().toString().padStart(2, "0");
  const hour   = date_time.getHours().toString().padStart(2, "0");
  const minute = date_time.getMinutes().toString().padStart(2, "0");
  return `${year}-${month}-${date} ${hour}:${minute}`;
}

function sizeString(bytes) {
  if(bytes < 1024) { return `${bytes} B`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} KiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} MiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} GiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} TiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} PiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} EiB`; }
  bytes /= 1024;
  if(bytes < 1024) { return `${bytes.toFixed(2)} ZiB`; }
  bytes /= 1024;
  return `${bytes.toFixed(2)} YiB`;
}

async function api(path, method, requestBodyJson=null) {

  const fetchOptions = { method:method };
  if(requestBodyJson) {
    fetchOptions.headers = { "Content-Type": "application/json" };
    fetchOptions.body = JSON.stringify(requestBodyJson);
  }
  const fetchResult = await fetch(`${path}`, fetchOptions);
  const status = fetchResult.status;
  const body   = await fetchResult.json();
  return { status:status, body:body };
}

async function apiFile(path, method, fileName, fileData) {
  const formData = new FormData();
  formData.append(fileName, fileData);

  const fetchOptions = { method:method };
  fetchOptions.body = formData;
  const fetchResult = await fetch(`${path}`, fetchOptions);
  const status = fetchResult.status;
  const body   = await fetchResult.json();
  return { status:status, body:body };
}

export { dateString, timeString, sizeString, apiFile };
export default api;
