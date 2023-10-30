function api(path, method, requestBodyJson, hasBody=false, statusThrowsError=true) {
  const fetchOptions = { method:method };
  if(requestBodyJson) {
    fetchOptions.headers = { "Content-Type": "application/json" };
    fetchOptions.body = JSON.stringify(requestBodyJson);
  }
  let fetchResult = fetch(`/admin/${path}`, fetchOptions);

  if(hasBody && statusThrowsError) {
    fetchResult = fetchResult.then((response) => {
      if((response.status < 200) || (response.status > 299)) { throw `${response.status}`; }
      return response.json();
    });
  } else if(hasBody) {
    fetchResult = fetchResult.then((response) => {
      return response.json();
    });
  } else if(statusThrowsError) {
    fetchResult = fetchResult.then((response) => {
      if((response.status < 200) || (response.status > 299)) { throw `${response.status}`; }
    });
  }
  
  return fetchResult;
}

export default api;
