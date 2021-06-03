import axios from 'axios';
import axiosRetry from 'axios-retry';
import * as qs from 'qs';

axiosRetry(axios, { retries: 3 });

export const API_CONFIG = {
	returnRejectedPromiseOnError: true,
	timeout: 30000,
	paramsSerializer: params => qs.stringify(params, { indices: false }),
	headers: {
		common: {
			'Content-Encoding': 'gz',
			'Cache-Control': 'no-cache, no-store, must-revalidate',
			'Content-Type': 'application/json',
			Accept: 'application/json',
			Pragma: 'no-cache',
		},
	},
};

// Backend api proxy instance
export const RootService = axios.create({
	...API_CONFIG,
	baseURL: process.env.DOCKER_API_URL,
});

// Backend api proxy instance
export const ApiService = axios.create({
	...API_CONFIG,
	baseURL: process.env.API_URL,
});

// Prometheus api proxy instance
export const PrometheusApiService = axios.create({
	timeout: 5000,
	headers: {
		common: {
			Pragma: 'no-cache',
		},
	},
	baseURL: process.env.METRICS_API_URL,
});
