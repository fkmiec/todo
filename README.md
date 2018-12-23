# Todolist

This is a fork of the Todolist project by Grant Ammons. I'm a big fan of TaskWarrior, but wanted a simpler cross-platform solution and wanted to learn Golang. Most of the work here follows the lead of TaskWarrior. 

Todolist is a simple and very fast task manager for the command line.

## Documentation
### Usage

<img src="https://github.com/fkmiec/todolist/blob/master/markdown/images/overall_help_commands.png" width="100%"/> 
<img src="https://github.com/fkmiec/todolist/blob/master/markdown/images/dates_filters_mods_args.png" width="100%" />  

### Basic Tutorial

First, create a command alias for 'todolist' to something shorter and easier to type like 'td'. From here on, we'll just refer to 'td'. 

alias td = todolist

### Create a Todo repository (Create in current directory. If created in home directory, will look there absent a repo in the local directory.)
$ td init  
Todo repo initialized.  

The repo consists of the following three files:
1) .todorc -- The config file for specifying defined reports, alias commands and named views
2) .todos.json -- The file containing all your current todos in JSON format
3) .todos_archive.json -- All archived todos
4) .todos_backlog.json -- Backlog of changes that will be used for syncing Todos with a remote server

### Add a Todo
$ td a My first todo  
Todo 1 added.

### Add a Todo for a project
$ td a My second todo +Project-1  
Todo 2 added.  

### Add a Todo for a specific context
$ td a My third todo @Home  
Todo 3 added.  

### Add a Todo with a due date
$ td a My fourth todo due:tomorrow  
Todo 4 added.  

### Add a Todo with project, context, due date
$ td a My fifth todo +Project-1 @Home due:tom  
Todo 5 added.  

### List your todos
![Example 1](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex1.PNG "Example 1")

### Filter your list to todos with the Home context
![Example 2](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex2.PNG "Example 2")

### Filter your list to todos with project Project-1
![Example 3](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex3.PNG "Example 3")

### Filter your todos to those with a due date
![Example 4](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex4.PNG "Example 4")

### Filter to the top 2 items for each defined project (Note I created a sixth todo with a separate project for this example)
#### You can also filter for the top N todos for a context (e.g. top:ctx:1)
![Example 5](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex5.PNG "Example 5")

### Define a custom report in the .todorc file
#### This example is a report sorted by id descending. You can, of course, show fewer columns, sort multiple and add filters
![Example 6](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex6.PNG "Example 6")

![Example 7](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex7.PNG "Example 7")

### Define aliases in .todorc to save typing on common commands
#### This example provides a short hand for the top 2 per project filter
![Example 8](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex8.PNG "Example 8")

![Example 9](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex9.PNG "Example 9")

### Define named view filters in .todorc that can be applied by default when listing todos (ie. 'td' or 'td list' or 'td default')
#### Below are defined two views. One for the Home context. The other for "work", which is defined as anything that isn't Home (ie. "minus" -@Home)
![Example 10](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex10.PNG "Example 10")

![Example 11](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex11.PNG "Example 11")

![Example 12](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex12.PNG "Example 12")

![Example 13](https://github.com/fkmiec/todolist/blob/master/markdown/images/ex13.PNG "Example 13")

### Ordinals for "all" todos and each project and context
Added support for setting an ordinal value for each todo relative to:
1) All todos ("all")
2) Every project
3) Every context

This adds capability to order (some might say prioritize) todos manually by setting the ordinal values. Rather than set specific ordinals, however, ordinals are set relative to the other todos. By default, each todo is given the ordinal for the last item in the set (all todos or a specific project or context). However, there is a command to (re)order todos using a relative ordering syntax.

td ord all:3,5,1  //Moves todos with ids 5 and 1 to just below id 3  
td ord all:0,8 //Move todo 8 to the first item in the set (0 is a special id that means top of the set)  
td ord +one:0,1,2,3 //Move 1,2 and 3 to top of set for project "one"  
td ord @home:4,1,9 //Move 1 and 9 behind 4 in the context for "home"  

Note that other todos not included in the comma-separated list of ids given to the order command are left in their current ordinal positions or shifted accordingly based on movement of the given ids. It takes some time to see how it works, but is flexible for positioning tasks one at a time or en masse. 

Sorting is supported by specifying the column as:

1) ord:all  //sort using ordinal values used for set of all todos  
2) ord:pro //sort using ordinal values for first project on the todo  
3) ord:ctx //sort using ordinal values for first context on the todo  

For project and context, sorting only considers the first project or context. All projects and contexts will have ordinals assigned, but there is no practical way to sort a single todo item with multiple projects by more than one at a time, so sort assumes the first project (or context).   

### Syncing repository with another file location
The sync command will synchronize your local repository with another file location. It performs a proper merge with support for adding, editing, deleting, completing, archiving, etc. 

A sync server was considered and rejected as unduly complex. If you have a cloud drive, you can sync your repository to that folder from multiple computers with no need to run a server or depend on someone else to do so. 

Encryption of the sync backlog file in the alternate file location is supported. Configure the following in your .todorc file:

sync.encrypt.passphrase=[passphrase | blank (don't encrypt) | * (prompt on cmd line)]  
sync.filepath=[filepath, including the actual filename, to sync your backlog to]  

The local backlog file will contain a UUID identifying the last sync. That UUID is maintained in the remote sync backlog to support identifying changes required to be merged. If you delete your local .todos_backlog.json file, the next sync will pull all todos from the remote sync backlog. Thus, syncing often serves as a backup and restore capability. 

### More details on filtering, sorting and applying due dates using relative date values like today, tomorrow, 1d, 5d, 1w, 1m, etc.
TBD


## License

Todolist is open source, and uses the [MIT license](https://github.com/fkmiec/todolist/blob/master/LICENSE.md).

