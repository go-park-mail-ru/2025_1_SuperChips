#!/usr/bin/env python3

import os
import requests
from requests_toolbelt.multipart.encoder import MultipartEncoder
import mimetypes

auth_url = "https://yourflow.ru/api/v1/auth/login"
upload_url = "https://yourflow.ru/api/v1/flows"
image_directory = "./static/img/"

print("please write yourflow.ru email and password")
email = input("email: ")
password = input("password: ")
credentials = {
    "email": email,
    "password": password
}

def authenticate():
    print("Authenticating...")
    response = requests.post(auth_url, json=credentials, verify=False)
    
    if response.status_code != 200:
        print(f"Authentication failed with status code {response.status_code}")
        print(response.text)
        exit(1)

    try:
        data = response.json().get("data")
        if not data:
            raise ValueError("data not found in response")
        
        csrf_token = data.get("csrf_token", "")
        if not csrf_token:
            raise ValueError("csrf token not found in response")
    except ValueError as e:
        print(e)
        exit(1)

    cookies = response.cookies
    if not cookies:
        print("No cookies received during authentication")
        exit(1)

    print("Authenticated successfully!")
    return csrf_token, cookies

def upload_images(csrf_token, cookies):
    print("Uploading images...")
    
    image_files = [f for f in os.listdir(image_directory) if os.path.isfile(os.path.join(image_directory, f))]
    if not image_files:
        print("No images found in the specified directory")
        return

    for image_file in image_files:
        print(image_file)
        file_path = os.path.join(image_directory, image_file)
        
        mime_type, _ = mimetypes.guess_type(file_path)
        if not mime_type:
            print(f"Could not determine MIME type for {image_file}. Skipping...")
            continue

        with open(file_path, "rb") as image:
            multipart_data = MultipartEncoder(
                fields={
                    "header": "some_header_value",
                    "image": (image_file, image, mime_type)
                }
            )

            headers = {
                "X-CSRF-TOKEN": csrf_token,
                "Content-Type": multipart_data.content_type
            }

            response = requests.post(upload_url, data=multipart_data, headers=headers, cookies=cookies, verify=False)

            if response.status_code == 200:
                print(f"Uploaded {image_file} successfully!")
            else:
                print(f"Failed to upload {image_file}. Status code: {response.status_code}")
                print(response.text)

if __name__ == "__main__":
    csrf_token, cookies = authenticate()

    upload_images(csrf_token, cookies)

