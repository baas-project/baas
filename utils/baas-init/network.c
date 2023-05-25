/*
 * Networking code for the BAAS init system which allows for lazy access to
 * other disk images. This could be implemented using a straight socket API
 * for easier porting, but right now this is a prototype of the concept.
 */

#include <curl/curl.h>
#include <string.h>

// Thread-local handle so we hide the abstraction
__thread CURL *curl_handle;

typedef size_t (*wb)(void *ptr, size_t size, size_t memb, void *buf);

size_t write_buffer(void *ptr, size_t size, size_t nmemb, void *buf) {
	size_t realsize = size * nmemb;
	memcpy(buf, ptr, realsize);
	return realsize;
}

size_t write_data(void *ptr, size_t size, size_t nmemb, void *stream) {
	size_t written = fwrite(ptr, size, nmemb, (FILE*) stream);
	return written;
}

void network_init() {
	curl_global_init(CURL_GLOBAL_ALL);
	curl_handle = curl_easy_init();
	/* curl_easy_setopt(curl_handle, CURLOPT_VERBOSE, 1L); */
	/* curl_easy_setopt(curl_handle, CURLOPT_NOPROGRESS, 1L); */

}

void network_add_standard_headers() {
	struct curl_slist *headers = NULL;
	headers = curl_slist_append(headers, "Origin: http://localhost:9090");
	headers = curl_slist_append(headers, "Type: system");
	curl_easy_setopt(curl_handle, CURLOPT_HTTPHEADER, headers);
}

void network_execute(void *target, wb cb) {
	network_add_standard_headers();
	curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, cb);
	curl_easy_setopt(curl_handle, CURLOPT_WRITEDATA, target);
	curl_easy_perform(curl_handle);
}

void network_set_uri(const char *uri) {
	curl_easy_setopt(curl_handle, CURLOPT_URL, uri);
}

void network_destroy() {
	curl_easy_cleanup(curl_handle);
	curl_global_cleanup();
}
