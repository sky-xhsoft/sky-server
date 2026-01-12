import requests
import json

url = "http://localhost:9090/api/v1/data/sys_table"
headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsImNvbXBhbnlJZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsImNsaWVudFR5cGUiOiJ3ZWIiLCJkZXZpY2VJZCI6IjNkZDg4ZDFhLWM5MTEtNGM5Yi1hOWI3LWUyMTkxNWY5ZDA3NCIsImV4cCI6MTc2ODE2MzI4MSwibmJmIjoxNzY4MTU2MDgxLCJpYXQiOjE3NjgxNTYwODF9.gDzeCd0UT3q6zPFJMRJ8glJxs3N6KkGLyFGK5cAwim0"
}
data = {
    "NAME": "TEST_API_CHECK",
    "DISPLAY_NAME": "测试",
    "DESCRIPTION": "test",
    "MASK": "AMDSQPGU",
    "SYS_TABLECATEGORY_ID": 1,
    "IS_ACTIVE": "Y",
    "ORDERNO": 9999
}

response = requests.post(url, json=data, headers=headers)
print(f"Status Code: {response.status_code}")
print(f"Response: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
