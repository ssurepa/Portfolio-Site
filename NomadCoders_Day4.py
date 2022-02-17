import os
import requests

try:
  while True:
    print("Welcome to IsItDown.py!\nPlease write a URL or URLs you want to check. (separated by comma)")
    split_urls = input().split(',')
    split_urls = [x.strip(' ') for x in split_urls]
    
    for url in split_urls:
      url = url.lower()
      
      if url.startswith('http://'):
        clean_url = url
      else:
        clean_url = "http://" + url
      
      if "." in url:
        try:
          r = requests.get(clean_url)
          if r.status_code == 200:
            print(f"{clean_url} is up!")
          else:
            print(f"{clean_url} is down!")
        except:
          print(f"{clean_url} is down!")
      else:
        print(f"{url} is not a valid URL")    
    

    start_over = []
    while start_over != "n" and start_over != "y":
      print("Do you want to start over? y/n")
      start_over = input().lower()
      if start_over == "n":
        print("k. bye!")
        exit()
      elif start_over =="y":
        clear = lambda: os.system('clear')
        clear()
        break
      else:
        print("That's not a valid answer")
except:
  pass    