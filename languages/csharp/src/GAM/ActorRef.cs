﻿// //-----------------------------------------------------------------------
// // <copyright file="ActorRef.cs" company="Asynkron HB">
// //     Copyright (C) 2015-2016 Asynkron HB All rights reserved
// // </copyright>
// //-----------------------------------------------------------------------

using System.Threading.Tasks;

namespace GAM
{
    public class PID
    {
        public string Id { get; set; }

        public void Tell(object message)
        {
            var reff = ProcessRegistry.Instance.Get(this);
            reff.Tell(message);
        }

        public void SendSystemMessage(SystemMessage sys)
        {
            var reff = ProcessRegistry.Instance.Get(this);
            reff.SendSystemMessage(this, sys);
        }

        public void Request(object message, PID sender)
        {
            Tell(new Request(message, sender));
        }

        public async Task<T> RequestAsync<T>(object message)
        {
            var tsc = new TaskCompletionSource<T>();
            var p = Actor.FromProducer(() => new FutureActor<T>(tsc));
            var fpid = Actor.Spawn(p);
            Tell(new Request(message, fpid));
            await tsc.Task;
            return tsc.Task.Result;
        }

        public void Stop()
        {
            var reff = ProcessRegistry.Instance.Get(this);
            reff.Stop(this);
        }
    }

    public class Request
    {
        public Request(object message, PID sender)
        {
            Message = message;
            Sender = sender;
        }

        public object Message { get; }
        public PID Sender { get; }
    }

    public abstract class ActorRef
    {
        public abstract void Tell(object message);

        public void Stop(PID pid)
        {
            SendSystemMessage(pid, new Stop());
        }

        public abstract void SendSystemMessage(PID pid, SystemMessage sys);
    }

    public class LocalActorRef : ActorRef
    {
        public LocalActorRef(IMailbox mailbox)
        {
            Mailbox = mailbox;
        }

        public IMailbox Mailbox { get; }

        public override void Tell(object message)
        {
            Mailbox.PostUserMessage(message);
        }

        public override void SendSystemMessage(PID pid, SystemMessage sys)
        {
            Mailbox.PostSystemMessage(sys);
        }
    }
}