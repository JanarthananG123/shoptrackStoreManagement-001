import axios from 'axios';

const gBaseURL = 'http://localhost:8080/api';

const gApiClient = axios.create({
  baseURL: gBaseURL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

export default gApiClient;
