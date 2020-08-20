- Every .go file in the same directory belongs to the same package. No exceptions. If you split your code into two separate directories, they are two separate packages now. Packages should have the same name as the directory that they reside in.
    - For a crash course on idiomatic package naming in go, refer to https://blog.golang.org/package-names
- All .go files in a package are treated like they are concatenated into a single .go file under that package. So if you declare a variable in one .go file in the a package, you can magically access it in another .go file under the same package. This means filenames are meaningless to the go compiler and you can freely rename files and move code blocks into other files under the same package without any impact on compilation.
- Package hierarchy. Go tolerates no cyclic imports.
    - Under the the app package contains every role with its own separate package, and every role package imports the skylab package. So skylab cannot import any other role package. It is not advisable to import a role package in another role package e.g. code in admin package should not import student package. All role packages should be separate, standalone silos. This means that if a role package fails, the other role packages will not be affected since they only depend on the skylab package.
- Subpackages are treated as separate packages from their parent package. So although the admin package is nested inside the app package, it is treated like any other external package. You could move the admin directory out of the app directory and nothing would change (except for the import paths).
- Receiver functions are just like normal OOP object methods, while structs are like OOP objects. https://yourbasic.org/golang/methods-explained/
```java
// Java
public class Dog {
   String breed;
   int age;
   String color;

   public Dog(String breed, int age, String color) {
       this.breed = breed;
       this.age = age;
       this.color = color
   }
   public string bark() { return "woof I am a "  + breed; }
   public void sleep() { }
}
// Dog dog = new Dog("Shiba Inu", 28, "yellow");
// dog.bark(); --> "woof I am a Shiba Inu"
// dog.sleep();
```
```go
// Go
type Dog struct {
    breed string
    age int
    color string
}
func (dog Dog) bark() string { return "woof I am a " + dog.breed }
func (dog Dog) sleep() { }
// dog := Dog{breed: "Shiba Inu", age: 28, color: "yellow"}
// dog.bark() --> "woof I am a Shiba Inu"
// dog.sleep()
```
- Interfaces are how different types can act as the same type (or interface, to be more pedantic). Getting a type to be treated as an interface is simply a matter of declaring all the methods on the type that the interface declares. https://yourbasic.org/golang/interfaces-explained/
```go
type Person interface {
    Name() string
    Age() int
    Speak(string) string
}

type A int
func (a A) Name() string { return "a" }
// type A is not a Person
// a.Name() --> "a"
// a.Age() --> ERROR! no such method 'Age'
// a.Speak("hello") --> ERROR! no such method 'Speak'

type B int
func (b B) Name() string { return "b" }
func (b B) Age() int { return 8 }
// type B is not a Person
// b.Name() --> "b"
// b.Age() --> 8
// b.Speak("hello") --> ERROR! no such method 'Speak'

type C int
func (c C) Name() string { return "c" }
func (c C) Age() int { return 8 }
func (c C) Speak(words string) string { return words }
// Success! Type C is a person
// c.Name() --> "c"
// c.Age() --> 8
// c.Speak("hello") --> "hello"
```
- Pages have filenames in their header tag. Use this to locate which template is responsible for generating the page. Most helpfully the go function that calls the template is usually called the same filename as the html file e.g. if you want to see which function is calling `skylab/admin/dashboard.html`, look inside the `skylab/admin/dashboard.go` file.
- Use `dumpjson=true` query parameter for introspecting into a page's data that was used to populate the template.
- Debug Mode is your friend. When enabled (by default), it will print all the handler functions that a request passes through.
    - Extremely illuminating for anyone trying to understand how the codebase works.
    - The handler functions associated with the URLs are defined in routes.go. That is where you can add or remove handlers. Every role package (admin, student, adviser etc) has its own routes.go file.
    - Unfortunately routes.go is not the most readable as it eschews URL string literals for Sections (defined in section.go) so you never really know what the full URL is without doing a section lookup. This was done so that the URL of a section can be changed in one place (in section.go) and be changed everywhere. To compensate for route.go's unreadability, the `[TRACE]` logging statements act as a shortcut to routes.go as it will print the function, filename and line number of the handler(s) visited without needing to consult routes.go.
    - The trace statements are manually inserted. Every handler should start with a `Skylab.TraceRequest()` function call which is responsible for printing the `[TRACE]` logging statements.
- Use Skylab.Log.Printf/Println as quick and easy printing to stdout. They will show you which file and line number they were invoked from, making it very useful to see which line in the output corresponds to which Printf/Println function call.
- Hover over a sidebar item to see what its section is. // Outdated
- Use Skylab.Log.SqlPrintf(). It takes the exact same arguments as Exec/Queryx/QueryRowx and will interpolate the arguments into the query and print a `[DEBUG]` logging statement together with the filename and line number where it was called from. Again, it makes monitoring a request and the database calls that it invokes much easier to trace.
- End handlers (the handlers that are resposible for rendering the template) have type signature `http.HandlerFunc`. Middleware handlers have type signature `func(http.Handler) http.Handler`.
- type `http.HandlerFunc` is the basic handler signature to use. `http.Handler` is just `http.HandlerFunc` except you must call `http.Handler.ServeHTTP` instead. Calling `http.Handler.ServeHTTP` is just like calling `http.HandlerFunc`. The reason why `http.Handler` is used is because that is what we must use for middlewares. That's just the convention.
    - The type signature `func(http.ResponseWriter, *http.Request)` is considered a `http.HandlerFunc`, but is not considered a `http.Handler`. To typecast (called a type conversion in Go) a `func(http.ResponseWriter, *http.Request)` to a `http.Handler` you must wrap it in a `http.HandlerFunc` type conversion i.e. `http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ })`.
        - A type conversion is the same way a slice of characters (bytes) can be converted as a string e.g. `name := string([]byte{'j', 'a', 'c', 'k'}) // "jack"`
        - You may wonder this is illogical, why don't I wrap `func(http.ResponseWriter, *http.Request)` in a `http.Handler` call to type convert it into a `http.Handler`? Why must I wrap it in a `http.HandlerFunc` type conversion instead, when the type signature `func(http.ResponseWriter, *http.Request)` is already considered a `http.HandlerFunc`?
        - The reason is because `http.Handler` is not a type, it is an interface. All `http.HandlerFunc`s are also `http.Handler`s because it satisfies the `http.Handler` interface, but not all `http.Handler`s are `http.HandlerFunc`s because other types may be a `http.Handler` interface.
        - Instead of manually declaring a `.ServeHTTP()` method on `func(http.ResponseWriter, *http.Request)` to satisfy the `http.Handler` interface, you can just type convert your `func(http.ResponseWriter, *http.Request)` into a `http.HandlerFunc` and `.ServeHTTP()` is already declared on all `http.HandlerFunc`s.
        - Thus `func(http.ResponseWriter, *http.Request)` is not a `http.Handler`, but `http.HandlerFunc(func(http.ResponseWriter, *http.Request))` is a `http.Handler`.
        - Thanks for coming to my TED talk.
- If you find that certain requests are bypassing your handlers (i.e. you check the [TRACE] logs and the browser never even calls particular handler(s)) it may be due to the web browser caching the redirect page. If it sees that GET `localhost:8080/a` redirected to `localhost:8080/b` which redirected to `localhost:8080/c`, in the future whenever it sees a link to `localhost:8080/a` it will just redirect the user to `localhost:8080/c` directly, bypassing `localhost:8080/b` and any other handlers along the way. The solution is to tell the browser not to cache the result of visiting `localhost:8080/c` with the skylab.DoNotBrowserCache() function. Look it up in the codebase to see where and how it is used.
    - The browser also displays cached pages when a user hits the back button so that it does not make another network request. For some pages such as listing pages this may lead to displaying stale data i.e. the user submits his submission and presses back but the listing page still shows he has not submitted (until he refreshes the page). To prevent back page caching, also use `headerutil.DoNotCache(w)`.
- Cookies only 'exist' at the end of the request. Context only exists from start to the end of the request. If you set a cookie at the start of a request, do not be surprised when you find out you cannot access that cookie at some later point in the request.
```
    A() -> B() -> C() -> User -> ...
    └──────────────┘└──────────────▶
        Context        Cookies
```
Let's say the route "/home" will call up functions A(), B() and C() in order. If A() wants to pass some data down to C(), say the user's email, it can't set the data in a cookie and expect C() to read the cookie. Cookies only get written at the end of the request which is at C(), right before it writes HTML out to the user. So it has to write the data into context for C() to pick up. Conversely, context cannot persist across requests. Every redirect counts as a separate request. If A() wants to persist the user's email across a redirect, it has to write it into a cookie.
- If you get a mysterious runtime panic and you were dealing with maps before the panic, it is highly likely you tried to set a value in a map without first using make() to instantiate it. Check all occurrences of your map usage to see if you successfully instantiated all maps before using.
- If you've written to http.ResponseWriter already, you cannot append any more headers to the response. This is because the headers are the first thing you send to the user before you can send the rest of the body. Once you send the body, you can't send the headers anymore. It's just how the HTTP protocol works.
- You can write custom Go functions that you can call in html templates. This project leverages template functions very heavily, look into any html file and you will find references to functions that are not part of the standard template syntax e.g. {{SkylabCsrfToken}}. Functions are injected into a template by passing in a map of template functions to Skylab.Render. Skylab.Render itself will inject a bunch of template functions which is how you get some globally available template functions.
- All html forms with `method="post"` need a {{SkylabCsrfToken}} inside, which will expand into a `<input type=hidden ...>` containing a CSRF token. This is to defend against CSRF attacks.
    - If gorilla CSRF keeps mysteriously chimping out on you ('Invalid CSRF Token') even though you definitely included {{SkylabCsrfToken}} in the form, try rebooting your computer and clearing all history/cookies/etc. I don't know why it happens, and sometimes rebooting doesn't even work.
- If you find yourself writing what seems to be a common utility function (like finding the max integer in an integer slice), resist the temptation to dump it into utils.go and be done with it. The fewer the dependencies the better, each function should only have the least amount of visibility it needs. Instead, if you are ever going to use that function in only that one place (so far), declare the utility function within the function that you are going to use it in. If you are going to use the utility function in multiple functions in the same package, then declare it as an unexported function within the package. Only if you are using that utility function within multiple package then you should promote that utility function into a utils.go file. // may be outdated
    - For relevant reading see https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common
- If your js/ts imports mysteriously break when using webpack, check if you included the static/vendor.js file inside the html file. Every webpack-compiled js file needs the accompanying vendor.js file in order to work, because all the library code that you import resides in vendor.js.
- (Something about how sections are automatically added as global template functions)
- `expected 'package', found 'EOF'` is caused when you have a source file that doesn't start with `package <package_name>`. This can commonly happen when creating a new .go file that is totally empty. Simply having a totally empty go file without a package declaration is enough to break compilation. All go files must contain a package declaration.
- Right now the 'updated\_at' and 'deleted\_at' columns in various database tables are unused; updated\_at is initialised once at the start together with the 'created\_at' column and is never updated again. deleted\_at is never used at all. The point is that they're there and you don't need to go through a schema migration in order to add them in. They're in there and unused.
- If you need to add a new javascript/css CDN link e.g. stackpath.bootstrapcdn.com, you need to whitelist the domain in skylab.SecureHeaders (under 'Content-Security-Policy').
- All sql views start with v_ and same with their filenames. If no v_, it is an sql function.
- use loadsql for loading specific sql files(s). If no file is provided, it loads all valid function/view files.
- How do I set flash messages?
- What is SetRoleSectionCtx/ Why is the appropriate sidebar section not lighting up?
    - Why is the sidebar missing?
- I did try to implement a template caching system (saving templates into a map based on a hash of the list of the filenames) but it led to some stale template issues (flash messages not disappearing etc) so I scrapped the idea. If re-computing templates proves to be slowing down site performance, maybe take another look at it.
- RE: the media table. I know storing images in the database is not reccomended, but I needed a solution which does not rely on external providers because they have a very high likelihood of failing to be maintained (who is paying for the S3 buckets? is everyone familiar with how to use AWS's/some other cloud storage provider's API?)
    - Instead all media access is abstracted behind a uuid. Every piece of media is identified and is retrievable by its uuid. The uuid lookup can be done in an S3 bucket in the future, but right now it's done by searching through the media table in the database.
    - My point is it's easy to migrate to S3 because all media access is done through a uuid anyway, which is how S3 stores media (behind a key and value pair).
    - If S3 doesn't work out, we can always fall back to the database storage solution (which works and needs no external dependencies).
- If you encounter some sql function starting with 'app.\<something\>' and don't know where it came from or what it does, know that all functions under the 'app' schema are defined in the app/db/functions/ directory.
    - If you want to know how the functions in the app/db/functions directory end up in the database, they are inserted into the database when you run cmd/loadsql/main.go (or its compiled executable). Specifically, look at the sortdirsV2() function in cmd/loadsql/main.go for how it does it.
- what the business with the 'pseudo-null' foreign keys are about
- for the form\_schema table:
    - 'name' column differentiates between different forms under the same period
    - 'subsection' column differentiates between different sections under the same form
- A project-wide searching tool in extremely invaluable for navigating the codebase.
    - If you see an unknown template function in some HTML file, you don't know what it does or where it was defined. This is where project-wide searching of the function name will come in handy.
    - You see some error being thrown in the console or the browser page and you don't know where it's coming from. You can project-wide search the error string in order to pretty reliably locate the source.
    - If you use Goland IDE, you already have this functionality built in. Simply go Edit > Find > Find In Path and enter the string you're searching for (https://www.jetbrains.com/help/idea/finding-and-replacing-text-in-project.html).
    - If you've decided to use your favourite text editor, you can use an external CLI tool for project-wide search instead. In that case I recommend you use [ripgrep](https://github.com/BurntSushi/ripgrep), which is available on all platforms.
        - The simplest way to use ripgrep is `rg <your_search_term>` and ripgrep will recursively search all files in the current directory for that search term. 
        - There are more advanced usages of ripgrep which you can find online.
- "My javascript doesn't work and the console shows: `Refused to run the JavaScript URL because it violates the following Content Security Policy directive:` (or similar)"
    - Plain inline scripts `<script></script>` are disabled due to CSP. To allow inline scripts, use a nonce. [https://content-security-policy.com/examples/allow-inline-script/](https://content-security-policy.com/examples/allow-inline-script/)
    - The nonce is available in all templates via the template function 'HeadersCSPNonce'. Add it to your inline scripts like so `<script nonce="{{HeadersCSPNonce}}"></script>`.
