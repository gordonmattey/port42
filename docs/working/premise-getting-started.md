# Getting Started with Premise: Programming by Declaring What Should Be True

## What if programming was about describing reality, not instructing computers?

Most programming languages make you think like a machine. You write step-by-step instructions: fetch this, calculate that, loop here, return there. It's exhausting. You spend more time managing the computer than solving your problem.

Premise is different. Instead of telling the computer *how* to do things, you declare *what* should be true. The language figures out the how.

## Your First Taste: Hello, Reality

In traditional programming, you might write:
```python
# Python
users = []
def add_user(name, email):
    users.append({"name": name, "email": email})
    return len(users) - 1
```

In Premise, you declare what exists:
```premise
(relation User
  :Name
  :Email)

(new User :Name "Alice" :Email "alice@example.com")
```

That's it. No arrays to manage. No functions to write. You declared that Users exist and created one. The system handles storage, retrieval, and persistence automatically.

## The Mental Model Shift

### Traditional Programming: Procedural Instructions
```javascript
// JavaScript
let balance = 100;
function withdraw(amount) {
    if (balance >= amount) {
        balance -= amount;
        return true;
    }
    return false;
}
```

### Premise: Declarative Relationships
```premise
(relation Account
  :Balance 100)

(rule withdrawal
  when [Account :Balance as ?b]
       (>= ?b ?amount)
  do (put Account :Balance (- ?b ?amount)))
```

Notice the difference? In Premise, you're not writing a function. You're declaring a rule about reality: "When an account has sufficient balance, a withdrawal can occur." The system handles the execution.

## Core Concepts in 5 Minutes

### 1. Relations: Your Data Structure
Relations are like classes, but they automatically persist and can be queried:

```premise
(relation Task
  :Title
  :Status "pending"
  :AssignedTo)

(new Task :Title "Write documentation" :AssignedTo "Bob")
(new Task :Title "Review PR" :AssignedTo "Alice")
```

### 2. Pattern Matching: Query Your Reality
Find things by describing their shape:

```premise
; Find all pending tasks assigned to Alice
(with [Task :Status = "pending" :AssignedTo = "Alice"] 
  list ?task)
```

### 3. Rules: Reality That Maintains Itself
Rules automatically trigger when conditions are met:

```premise
(rule auto-assign
  when [Task :Status = "pending" :AssignedTo = nil]
  do (put Task :AssignedTo "DefaultWorker"))
```

### 4. Everything Is Connected
Unlike traditional programming where data lives in silos, everything in Premise exists in the same space:

```premise
(relation Project
  :Name
  :Tasks [])

(rule link-task-to-project
  when [Task ^ as ?t :Title as ?title]
       [Project :Name = "CurrentProject"]
  do (enq Project :Tasks ?t))
```

## A Real Example: Building a Simple Team Workspace

Let's build something useful - a team workspace that tracks tasks and discussions:

```premise
; Define our world
(relation Team
  :Name
  :Members [])

(relation Member  
  :Name
  :Role
  :JoinedAt (moment))

(relation Discussion
  :Topic
  :Author
  :Content
  :Timestamp (moment)
  :Replies [])

(relation Task
  :Title
  :Description
  :AssignedTo
  :Status "pending"
  :CreatedAt (moment))

; Create a team
(new Team :Name "SnowOS Builders")

; Add members
(new Member :Name "Sarah" :Role "Founder")
(new Member :Name "Chen" :Role "Engineer") 
(new Member :Name "Maria" :Role "Designer")

; Rules that maintain our workspace
(rule welcome-new-member
  when [Member ^ as ?m :Name as ?name :JoinedAt as ?time]
       (< (- (moment) ?time) 60)
  do (new Discussion 
       :Topic "Welcome!"
       :Content ($ "Welcome to the team, " ?name "!")
       :Author "System"))

(rule task-completion-cascade
  when [Task :Status = "completed" :Title as ?title]
  do (new Discussion
       :Topic "Task Completed"
       :Content ($ ?title " has been completed!")
       :Author "System"))

; Query our workspace
(with [Task :AssignedTo = "Chen" :Status = "pending"]
  list :Title)
; Returns all pending tasks for Chen

(with [Discussion :Timestamp > (- (moment) 86400)]
  tally)
; Count discussions in last 24 hours
```

## The Magic: No Database, No API, No Servers

Everything you just declared is automatically:
- **Persisted**: Survives restarts
- **Queryable**: Search by any attribute
- **Reactive**: Rules trigger automatically
- **Connected**: Relations reference each other naturally

You never wrote:
- Database schemas
- API endpoints  
- State management
- Event handlers
- Serialization logic

## Why This Matters

Traditional programming makes you build walls between:
- Code and data
- Client and server
- Memory and storage
- Present and past

Premise dissolves these walls. Your data isn't in a database - it's part of your program. Your rules aren't event handlers - they're declarations about reality. Your queries aren't SQL - they're patterns.

## When to Use Premise

Premise shines when you need:
- **Living systems** that maintain their own consistency
- **Complex relationships** between entities
- **Rules and patterns** rather than procedures
- **Persistent state** without database hassle
- **Exploratory programming** where requirements evolve

It's perfect for:
- AI agent systems (agents as relations with rules)
- Knowledge bases (facts that derive new facts)
- Workflow engines (rules that maintain state)
- Collaborative systems (shared reality across users)
- Game logic (entities with relationships and rules)

## Your Next Steps

1. **Think in Relations**: What entities exist in your domain?
2. **Declare, Don't Instruct**: What should be true?
3. **Write Rules**: What maintains consistency?
4. **Trust the System**: Let Premise handle the how

## The Paradigm Shift

Remember: You're not programming a computer. You're describing a reality. The computer figures out how to make it true.

In Premise, you don't write code that manipulates data. You declare relationships and rules, and reality maintains itself.

Welcome to programming without walls. Welcome to Premise.

---

*Ready to dive deeper? The best way to learn Premise is to start declaring your own reality. What world do you want to build?*