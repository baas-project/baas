# Control Server Logging
Users can be registered or logged only via the REST endpoint or by using the OAuth2 protocol. Currently only GitHub is supported as OAuth source, but more will be added when required or possible.

First send a GET request to `/user/login/github` which redirects you to the Github login page. After giving your permission, user data is exchanged and user is created. Here you can find your session ID as a cookie in the website inspector. This session ID can be sent to the server on future requests in order to authenticate yourself. There is functionally no difference between logging and registering since the user will be made on first login.

At the moment is not possible to register multiple OAuth sources to one account. Each login as seen as unique and distinct even with shared data. Keep in mind that if you use the same username for multiple sources it may conflict and reject the login.

## OAuth usage flow
```plantuml
|User|
start

:Send login request>

|LogingSource|
:Receive initial login request<
:Return the authentication URL>

|User|
:Receive URI<
:Authenticate on authentication page;

|LoginCallback|
:Get the data session and Github code<
:Fetch the OAuth token<
:Request the user data>
if (User not exists?) then (yes)
   :Create the user in the database;
endif


:Set Username in the session<
:Return session ID<

|User|
:Receive session ID<
:Send any arbitrary request>

|RequestHandler|
:Receive session>
if (Session not set?) then
  :Reject call and return error<
  end
endif

if (Can user role access it?) then (no)
   :Reject call and return error<
   end
endif

if (Is it about this user?) then (no)
   :Reject call and return error<
   end
endif

:Handle call and return response<
stop
```
