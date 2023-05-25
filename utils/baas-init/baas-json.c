#include "cJSON.h"
#include <stdbool.h>
#include <stdlib.h>

struct baas_image {
	char *name;
	int version;
	char *uuid;
	char *username;
	char *diskcompressionstrategy;
	char *imagefiletype;
	char *type;
	char *checksum;
	char *filesystem;
};

struct baas_image_frozen {
	struct baas_image *image;
	bool update;
};

struct baas_setup {
	struct baas_image_frozen **images;
	int images_len;
	char *name;
	char *username;
	char *uuid;
};


char *json_get_string(const char *name, const cJSON *json) {
	const cJSON *val = cJSON_GetObjectItemCaseSensitive(json, name);

	if (cJSON_IsString(val) && val->valuestring) {
		return val->valuestring;
	}

	return "";
}

int json_get_int(const char *name, const cJSON *json) {
	const cJSON *val = cJSON_GetObjectItemCaseSensitive(json, name);

	if (cJSON_IsNumber(val) && val->valueint) {
		return val->valueint;
	}

	// We assume that no version can be -9
	return 0;
}

struct baas_image *parse_baas_image(const cJSON *json) {
	struct baas_image *bi = malloc(sizeof(struct baas_image));
	bi->name = json_get_string("Name", json);
	bi->uuid = json_get_string("UUID", json);
	bi->diskcompressionstrategy = json_get_string("DiskCompressionStrategy", json);
	bi->imagefiletype = json_get_string("ImageFileType", json);
	bi->type = json_get_string("type", json);
	bi->checksum = json_get_string("Checksum", json);
	bi->filesystem = json_get_string("Filesystem", json);
	return bi;
}

struct baas_image_frozen *parse_baas_image_frozen(const cJSON *json) {
	struct baas_image_frozen *bif = malloc(sizeof(struct baas_image_frozen));
	bif->image = parse_baas_image(cJSON_GetObjectItemCaseSensitive(json, "Image"));

	int version = json_get_int("Version", cJSON_GetObjectItemCaseSensitive(json, "Version"));
	bif->image->version = version;

	const cJSON *jUpdate = cJSON_GetObjectItemCaseSensitive(json, "Update");
	bif->update = jUpdate->valueint;

	return bif;
}
struct baas_setup *parse_baas_setup(const cJSON *json) {
	struct baas_setup *bs = malloc(sizeof(struct baas_setup));
	const cJSON *images = cJSON_GetObjectItemCaseSensitive(json, "Images");

	// Create a list that can hold the
	bs->images_len = cJSON_GetArraySize(images);
	bs->images = calloc(sizeof(struct baas_image_frozen), bs->images_len);

	int i = 0;
	const cJSON *image;
	cJSON_ArrayForEach(image, images) {
		bs->images[i++] = parse_baas_image_frozen(image);
	}


	bs->name = json_get_string("Name", json);
	bs->username = json_get_string("Username", json);
	bs->uuid = json_get_string("UUID", json);
	return bs;
}

void free_baas_image(struct baas_image *bi) {
	free(bi);
}

void free_baas_image_frozen(struct baas_image_frozen *bif) {
	free(bif->image);
	free(bif);
}

void free_baas_setup(struct baas_setup *bs) {
	free(bs->images);
	free(bs);
}
