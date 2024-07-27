import axios from "axios";

const api = axios.create({
  baseURL: "http://localhost:8080", // Your backend URL
});

export const callApi = async (endpoint, data) => {
  await fetch(`http://localhost:8080${endpoint}`).then((res) => {
    return res;
  });
};

export const signUpUser = (data) => api.post("/signup", data);
export const loginUser = (data) => api.post("/login", data);
