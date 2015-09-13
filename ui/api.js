import fetch from 'isomorphic-fetch';
//import { Schema, arrayOf, normalize } from 'normalizr';

const API_URI = '/api';
const AUTH_TOKEN = 'X-Auth-Token';

function getToken() {
  return window.localStorage.getItem(AUTH_TOKEN);
}

function setToken(token) {
  if (token) {
    window.localStorage.setItem(AUTH_TOKEN, token);
  }
}

function deleteToken() {
  window.localStorage.removeItem(AUTH_TOKEN);
}

function callAPI(endpoint, method, data) {

  method = method || "GET";

  const args = { method: method };
  const token = getToken();

  let headers = {};

  if (token) {
    headers[AUTH_TOKEN] = token;
  }

  if (data) {
      // check if window.FormData
      if (data instanceof window.FormData) {
        args.body= data;
      } else {
        args.body  = JSON.stringify(data);
        headers = Object.assign({}, headers, {
          "Accept": "application/json",
          "Content-Type": "application/json"
        });
      }
  }

  if (headers) {
    args.headers = headers;
  }

  return fetch(API_URI + endpoint, args)
    .then(response => {
      if (response.headers.has(AUTH_TOKEN)) {
        const token = response.headers.get(AUTH_TOKEN);
        if (token) {
          setToken(token);
        }
      }
      if (response.headers.get('Content-Type').match('application/json')) {
        return response.json();
      }
      return response;
    });

}

export function getPhotos(page, orderBy) {
  return callAPI(`/photos/?page=${page}&orderBy=${orderBy}`);
}

export function searchPhotos(page, query) {
  return callAPI(`/photos/search?page=${page}&q=${query}`);
}

export function updatePhotoTitle(id, title) {
  return callAPI(`/photos/${id}/title`, 'PATCH', {
    title: title
  });
}

export function updatePhotoTags(id, tags) {
  return callAPI(`/photos/${id}/tags`, 'PATCH', {
    tags: tags
  });
}

export function getPhotoDetail(id) {
  return callAPI('/photos/' + id);
}

export function deletePhoto(id)  {
  return callAPI('/photos/' + id, 'DELETE');
}

export function getUser() {
  return callAPI('/auth/');
}

export function logout() {
  return callAPI('/auth/', 'DELETE').then(() => deleteToken());
}

export function login(identifier, password) {
  return callAPI('/auth/', 'POST', {
    identifier: identifier,
    password: password
  });
}

export function getTags() {
  return callAPI('/tags/');
}

export function upload(title, tags, photo) {
  const data = new window.FormData();
  data.append("photo", photo);
  data.append("title", title);
  data.append("tags", tags);

  return callAPI('/photos/', 'POST', data);

}

