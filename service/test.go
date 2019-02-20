package service

import (
	"cheat/logger"
	"cheat/model"
	"cheat/model/plate"
	"cheat/orm"
	"cheat/protocol"
	"cheat/util"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/phf/go-queue/queue"
	"sync"
	"sync/atomic"
	"time"
)

func Test(conn *websocket.Conn) {

	for {

		//读取
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Debug("msg type:", mt, "| msg data:", msg)

		//写
		err = conn.WriteMessage(mt, msg)

		if err != nil {
			logger.Error(err)
			return
		}
	}

}


type UserManager struct {
	sync.Map

	LogoutChan chan int64
}

var UserMap UserManager

type UserContext struct {
	Conn *websocket.Conn

	RecvChanMap map[protocol.EventType]chan []byte
	SendChanMap map[protocol.EventType]chan []byte

	Cancel context.CancelFunc
	RoomId int64
	Uid int64
}

var RoomManager sync.Map

func (u *UserContext) joinRoom(room *GameRoom) {

}
func (u *UserContext) exitRoom(room *GameRoom) {

}
func CreateRoom() *GameRoom {

	return nil
}
func (u *UserContext) processRoomMsg(rcmd *protocol.RoomRequest) {

	switch rcmd.Type {

	case protocol.RC_Nil:
		return

	case protocol.RC_Create:
		{
			rm := CreateRoom()
			rrc := protocol.RoomResCreate{
				Id:          rm.Id,
				Name:        rm.Name,
				Password:    rm.Pass,
				PlayerCount: len(rm.Player),
			}
			res := &protocol.RoomResponse{
				Type: protocol.RC_Create,
				Data: rrc,
			}
			u.SendRoomMsg(res)
		}
	case protocol.RC_Join:
		{
			room, exist := RoomManager.Load(rcmd.RoomId)
			if !exist {
				u.SendError(protocol.NewErr(protocol.ErrRoomNotExistCode))
				return
			}
			gr, ok := room.(*GameRoom)
			if !ok {
				logger.Error("room not exist!")
				return
			}
			u.joinRoom(gr)
		}
	case protocol.RC_Rank:
		{

		}
	case protocol.RC_List:
		{
			list := []protocol.RoomResListSingle{}
			RoomManager.Range(func(key, value interface{}) bool {
				tmp := protocol.RoomResListSingle{}
				list = append(list, tmp)
				return true
			})
			res := &protocol.RoomResponse{
				Type: protocol.RC_List,
				Data: list,
			}
			u.SendRoomMsg(res)
		}
	case protocol.RC_Exit:
		{
			//玩家不在房价 退出的房间号有问题 退出的房间号不等于玩家所在房价
			if u.RoomId == 0 || rcmd.RoomId == 0 || rcmd.RoomId != u.RoomId {
				return
			}
			room, exist := RoomManager.Load(rcmd.RoomId)
			if !exist {
				u.SendError(protocol.NewErr(protocol.ErrRoomNotExistCode))
				return
			}
			gr, ok := room.(*GameRoom)
			if !ok {
				logger.Error("room not exist!")
				return
			}
			u.exitRoom(gr)
		}
	}
}
func (u *UserContext) processPlayMsg(rcmd *protocol.PlayRequest) {

}
func (u *UserContext) processSignMsg(rcmd *protocol.SignRequest) {
	switch rcmd.Type {
	case protocol.SC_Login:{

	}
	case protocol.SC_Logout:{

	}
	default:{

	}
	}
}
func (u *UserContext) processChatMsg(rcmd *protocol.ChatRequest) {

}

func (u *UserContext) ProcessMsg(ctx context.Context) {

	for {

		select {
		case <-ctx.Done():
			return
		case rmsg := <-u.RecvChanMap[protocol.UC_RoomCmd]:
			{
				rcmd := &protocol.RoomRequest{}
				if err := json.Unmarshal(rmsg, rcmd); err != nil {
					logger.Warnning(err)
					continue
				}
				u.processRoomMsg(rcmd)
			}

		case rmsg := <-u.RecvChanMap[protocol.UC_PlayCmd]:
			{

				rcmd := &protocol.PlayRequest{}
				if err := json.Unmarshal(rmsg, rcmd); err != nil {
					logger.Warnning(err)
					continue
				}
				u.processPlayMsg(rcmd)

			}

		case rmsg := <-u.RecvChanMap[protocol.UC_ChatCmd]:
			{
				rcmd := &protocol.ChatRequest{}
				if err := json.Unmarshal(rmsg, rcmd); err != nil {
					logger.Warnning(err)
					continue
				}
				u.processChatMsg(rcmd)
			}

		case rmsg := <-u.RecvChanMap[protocol.UC_SignCmd]:
			{
				rcmd := &protocol.SignRequest{}
				if err := json.Unmarshal(rmsg, rcmd); err != nil {
					logger.Warnning(err)
					continue
				}
				u.processSignMsg(rcmd)
			}

		}

	}
}

func (u *UserContext) RecvMsg(ctx context.Context){
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				req := &protocol.WsRequest{}
				if err := u.Conn.ReadJSON(req); err != nil {
					logger.Warnning(err)
				}
				req.Uid=u.Uid
				switch req.Type {
				case protocol.UC_ErrCmd:
					continue
				case protocol.UC_SignCmd:
					u.RecvChanMap[protocol.UC_SignCmd] <- req.Data
				case protocol.UC_RoomCmd:
					u.RecvChanMap[protocol.UC_RoomCmd] <- req.Data
				case protocol.UC_PlayCmd:
					u.RecvChanMap[protocol.UC_PlayCmd] <- req.Data
				case protocol.UC_NotifyCmd:
					u.RecvChanMap[protocol.UC_NotifyCmd] <- req.Data
				}
			}
	}
	}
}
func (u *UserContext)SendMsg(ctx context.Context){
	for {
		res := &protocol.WsResponse{}
		select {
		case <-ctx.Done():
			return
		case data := <-u.SendChanMap[protocol.UC_ErrCmd]:
			{
				res.Type = protocol.UC_ErrCmd
				res.Data = data
			}
		case data := <-u.SendChanMap[protocol.UC_SignCmd]:
			{
				res.Type = protocol.UC_SignCmd
				res.Data = data
			}
		case data := <-u.SendChanMap[protocol.UC_RoomCmd]:
			{
				res.Type = protocol.UC_RoomCmd
				res.Data = data
			}
		case data := <-u.SendChanMap[protocol.UC_PlayCmd]:
			{
				res.Type = protocol.UC_PlayCmd
				res.Data = data
			}
		case data := <-u.SendChanMap[protocol.UC_NotifyCmd]:
			{
				res.Type = protocol.UC_NotifyCmd
				res.Data = data
			}
		}
	}
}

func (u *UserContext) SendPlayMsg(data []byte) {
	u.SendChanMap[protocol.UC_PlayCmd] <- data
}
func (u *UserContext) SendNotifyMsg(data []byte) {
	u.SendChanMap[protocol.UC_NotifyCmd] <- data
}
func (u *UserContext) SendRoomMsg(res *protocol.RoomResponse) {
	jstr, _ := json.Marshal(res)
	u.SendChanMap[protocol.UC_RoomCmd] <- jstr
}
func (u *UserContext) SendError(err error) {
	code := protocol.ErrUnDefineErrorCode
	if ie, ok := err.(protocol.InnerError); ok {
		code = ie.Code
	}
	er := &protocol.ErrorResponse{
		Code: code,
		Msg:  err.Error(),
	}
	jstr, _ := json.Marshal(er)
	u.SendChanMap[protocol.UC_ErrCmd] <- jstr
}
func (u *UserContext) ExitRoom() {
	pe := protocol.PlayRequest{
		PlayCmd: protocol.UPC_Exit,
	}
	jstr, _ := json.Marshal(pe)
	u.RecvChanMap[protocol.UC_PlayCmd] <- jstr
	u.RoomId = 0
}

type PlayStatus int

const (
	PS_Error      = iota
	PS_Ready      //准备
	PS_Wait_Plate //等待发牌
	PS_Current    //轮到
	PS_Wait_Cheat //等待下注
	PS_Win_All    //赢了
	PS_Lost       //输了
	PS_Exit       //离开
	PS_Warning    //警告
	PS_Close      //房间关闭
)

type RoomPlayer struct {
	Id     int64
	Amount uint64
	Status PlayStatus
	Watch  bool
	Plate  *plate.ThreePlate
}

type ServerMsg struct {
	//通知给玩家的状态
	Status PlayStatus

	//这人的id
	PlayerId int64

	//事件
	Event protocol.PlayRequest

	//需要通知的错误
	ErrStr string

	Plate *plate.ThreePlate

	//仅发送给某人
	SendId int64 `json:"-"`
}

const TimeOut = time.Second * 30

type RoomStatus int32

const (
	RS_Ready = iota
	RS_Cheat
)

type GameRoom struct {
	Id int64

	//房间名
	Name string
	//房间密码
	Pass string
	//房主
	Owner int64

	//房间用户列表
	Player map[int64]*RoomPlayer

	//桌上的人
	TablePlayer queue.Queue

	//用户命令的管道map
	PlayerChan chan *protocol.PlayRequest
	ServerChan chan *ServerMsg

	closeChanSend chan bool

	Timer *time.Timer

	Amount     uint64
	BaseAmount uint64
	CurVol     uint64
	CurUserId  int64

	Status   RoomStatus
	RoomLock int32
}

func (gr *GameRoom) generalPlate(count int) ([]*plate.ThreePlate,error) {
	res:=[]*plate.ThreePlate{}
	for i:=0;i<count;i++{

		tps:=plate.NewThreePlateSet()

		tpi:=tps.Get()
		if tpi==nil{
			logger.Error("plate set is empty")
			return nil,protocol.NewError("plate set is empty")
		}
		if tp,ok:=tpi.(*plate.ThreePlate);ok{
			res=append(res,tp)
		}else{
			logger.Error("general plate error")
			return nil,protocol.NewError("general plate error")
		}
	}
	return res,nil
}
func (gr *GameRoom) sendSmsg(msg *ServerMsg) {
	gr.ServerChan <- msg
}

func (gr *GameRoom) recvCmsg() *protocol.PlayRequest {
	return <-gr.PlayerChan
}

func (gr *GameRoom) recvUid(uid int64) {
	ctx, ok := getUser(uid)
	if !ok {
		logger.Warnning("not find user!", uid)
		return
	}
	for {
		jstr := <-ctx.RecvChanMap[protocol.UC_PlayCmd]
		event := protocol.PlayRequest{}
		if err := json.Unmarshal(jstr, &event); err != nil {
			logger.Warnning(err)
			continue
		}
		if event.PlayCmd == protocol.UPC_Error {
			//未知的指令
			continue
		}
		//检查发送过来的消息是否正确
		event.PlayerId = uid
		event.LoserId = 0
		event.Watch = gr.Player[uid].Watch
		//额度是否在限制内
		//event.Amount

		//查人是否在房间内
		if event.CheckId != 0 {
			if _, exist := gr.Player[event.CheckId]; !exist {
				//该玩家不在房间内
				continue
			}
		}

		if event.PlayCmd == protocol.UPC_Exit {
			//房间不再接收这人的消息
			gr.PlayerChan <- &event
			return
		}
		//压阶段
		if gr.CurUserId == 0 {
			event.Amount = gr.BaseAmount
			gr.PlayerChan <- &event
			continue
		}
		if uid != gr.CurUserId {
			//没轮到你发言
			continue
		}
		gr.PlayerChan <- &event
	}
}
func (gr *GameRoom) sendRoomPlayers() {
	for {
		select {
		case <-gr.closeChanSend:
			{
				return
			}
		case smsg := <-gr.ServerChan:
			{
				jst, err := json.Marshal(smsg)
				if err != nil {
					logger.Warnning(err)
				}
				//通知 给全部
				if smsg.SendId == 0 {
					for _, u := range gr.Player {
						if p, ok := getUser(u.Id); ok {
							p.SendPlayMsg(jst)
						}
					}
					continue
				}
				//通知 给指定用户
				if p, ok := getUser(smsg.SendId); ok {
					p.SendPlayMsg(jst)
				}
			}

		}
	}
}
func (gr *GameRoom) Close() {

	gr.closeChanSend <- true
	gr.sendSmsg(&ServerMsg{
		Status: PS_Close,
	})
	for _, p := range gr.Player {
		if player, exist := getUser(p.Id); exist {
			player.ExitRoom()
		}
	}
}

func (gr *GameRoom) Lock() {
	atomic.StoreInt32(&gr.RoomLock, 1)
}
func (gr *GameRoom) UnLock() {
	atomic.StoreInt32(&gr.RoomLock, 0)
}
func (gr *GameRoom) IsLock() bool {
	return atomic.LoadInt32(&gr.RoomLock) == 1
}
func (gr *GameRoom) checkExit() {
	//检查是否有退出的玩家,有就移除
	for _, p := range gr.Player {
		if p.Status == PS_Exit {
			delete(gr.Player, p.Id)
		}
	}
}

func (gr *GameRoom) play() {
	chect_func := func(msg *protocol.PlayRequest, u *RoomPlayer) error {
		//检查余额
		if msg.Amount > u.Amount {
			//没钱了
		}
		//检查金额

		//不看牌
		if u.Watch == false {
			if msg.Amount*2 < gr.CurVol {
				//太小了
			}
		} else { //看牌
			if msg.Amount < gr.CurVol {
				//太小了
			}
		}

		u.Amount -= msg.Amount

		tmpv := msg.Amount
		if u.Watch == false {
			tmpv = msg.Amount * 2
		}
		if tmpv < gr.CurVol {
			tmpv = gr.CurVol
		}

		gr.CurVol = tmpv

		gr.Amount += msg.Amount
		//下注成功
		return nil
	}
	check_func := func(msg *protocol.PlayRequest, u *RoomPlayer) error {

		if gr.Player[u.Id].Plate.Less(gr.Player[msg.CheckId].Plate) {
			return errors.New("less")
		}
		return nil
	}

	//设置所有人为 等待发牌,未看牌,
	for _, v := range gr.Player {
		v.Status = PS_Ready
		v.Watch = false
	}
	gr.sendSmsg(&ServerMsg{Status: PS_Ready})
	gr.Timer = time.NewTimer(TimeOut)
	for {
		select {

		case event := <-gr.PlayerChan:
			{
				omsg := &ServerMsg{}
				omsg.PlayerId = event.PlayerId
				omsg.Event.PlayCmd = event.PlayCmd
				if err := func() error {
					switch event.PlayCmd {

					case protocol.UPC_Cheat:
						{
							if err := chect_func(event, gr.Player[event.PlayerId]); err != nil {
								return err
							}
							gr.Player[event.PlayerId].Status = PS_Wait_Plate
							omsg.Event.Amount = event.Amount
							omsg.Event.Volume = gr.CurVol
						}

					case protocol.UPC_Exit:
						{
							gr.Player[event.PlayerId].Status = PS_Exit
						}
					default:
						{
							return protocol.NewErr(protocol.ErrPlayerCmdRejectCode)
						}
					}
					return nil
				}(); err != nil {
					omsg.Status = PS_Warning
					omsg.SendId = event.PlayerId
					gr.sendSmsg(omsg)
				} else {
					gr.sendSmsg(omsg)
				}
			}
		case <-gr.Timer.C:
			{

				break

			}

		}
	}

	for _, p := range gr.Player {
		if p.Status == PS_Wait_Plate {
			gr.TablePlayer.PushBack(p.Id)
		}
	}

	plates,generr:= gr.generalPlate(gr.TablePlayer.Len())
	if generr!=nil{
		//致命错误
		return
	}
	//发牌
	for i := 0; i < gr.TablePlayer.Len(); i++ {
		tmp := gr.TablePlayer.PopFront().(int64)
		gr.Player[tmp].Plate = plates[i]
	}

	for {

		curUserId := gr.TablePlayer.PopFront().(int64)
		curUser := gr.Player[curUserId]

		//已经输了
		if curUser.Status == PS_Lost {
			continue
		}

		if gr.TablePlayer.Len() == 0 {
			//赢了

			gr.sendSmsg(&ServerMsg{
				Status:   PS_Win_All,
				PlayerId: curUserId,
			})
			break
		}
		gr.Timer = time.NewTimer(TimeOut)

		//通知所有人 当前用户在干嘛
		gr.sendSmsg(&ServerMsg{
			Status:   PS_Current,
			PlayerId: curUserId,
		})

		if err := func() error {

			for {
				select {

				case event := <-gr.PlayerChan:
					{
						if event.PlayerId == curUserId {
							//当前用户的动作 转发给其他人
							omsg := &ServerMsg{}
							omsg.PlayerId = curUserId
							omsg.Event.PlayCmd = event.PlayCmd
							omsg.Status = PS_Wait_Cheat
							NeedNext := errors.New("needNext")
							if err := func() error {

								switch event.PlayCmd {
								case protocol.UPC_Cheat:
									{
										//押注

										if err := chect_func(event, curUser); err != nil {
											return err
										}
										omsg.Event.Amount = event.Amount
										omsg.Event.Volume = gr.CurVol
										omsg.Event.Watch = curUser.Watch

									}
								case protocol.UPC_Check:
									{
										//查人
										checkUser := gr.Player[event.CheckId]

										switch checkUser.Status {
										case PS_Lost:
											{
												//查牌的人已经挂了
												return protocol.NewErr(protocol.ErrPlayerLosedCode)
											}
										case PS_Wait_Plate:
											{
												//正常

												//查牌也要给钱的
												if err := chect_func(event, curUser); err != nil {
													return err
												}
												if err := check_func(event, curUser); err != nil {
													//比别人小
													gr.Player[curUserId].Status = PS_Lost
													omsg.Event.LoserId = curUserId
												} else {
													gr.Player[event.CheckId].Status = PS_Lost
													omsg.Event.LoserId = event.CheckId
												}
												omsg.Event.CheckId = event.CheckId
												omsg.Event.Amount = event.Amount
												omsg.Event.Volume = gr.CurVol
												omsg.Event.Watch = curUser.Watch

											}
										case PS_Exit:
											{
												return protocol.NewErr(protocol.ErrPlayerExitedCode)
											}
										default:
											{
												//异常
												return protocol.NewErr(protocol.ErrPlayerUnknowStatusCode)
											}
										}

									}
								case protocol.UPC_Watch:
									{
										//看牌
										watchMsg := &ServerMsg{
											Status: PS_Wait_Cheat,
											Event: protocol.PlayRequest{
												PlayCmd: protocol.UPC_Watch,
											},
											Plate:    curUser.Plate,
											PlayerId: curUser.Id,
											SendId:   curUser.Id,
										}
										gr.sendSmsg(watchMsg)
										//已经看牌
										gr.Player[curUserId].Watch = true

										//结束
										return NeedNext
									}
								case protocol.UPC_Throw:
									{
										//弃牌
										gr.Player[curUserId].Status = PS_Lost
										omsg.Status = PS_Lost
									}
								case protocol.UPC_Exit:
									{
										//退出
										gr.Player[curUserId].Status = PS_Exit
										omsg.Status = PS_Exit
									}
								default:
									{
										return protocol.NewErr(protocol.ErrPlayerCmdRejectCode)
									}
								}

								return nil
							}(); err != nil {
								//有错,只发给当前用户
								omsg.SendId = curUserId

								if err == NeedNext {
									gr.sendSmsg(omsg)
									continue
								}
								omsg.Status = PS_Warning
								omsg.ErrStr = err.Error()
								gr.sendSmsg(omsg)
							} else {
								//发送 给所有
								gr.sendSmsg(omsg)
							}
						} else {
							//不是当前人只能看和弃牌和退出
							switch event.PlayCmd {
							case protocol.UPC_Watch:
								{
									//看牌
									tuser := gr.Player[event.PlayerId]
									watchMsg := &ServerMsg{
										Status: PS_Wait_Cheat,
										Event: protocol.PlayRequest{
											PlayCmd: protocol.UPC_Watch,
										},
										Plate:    tuser.Plate,
										PlayerId: tuser.Id,
										SendId:   tuser.Id,
									}
									tuser.Watch = true
									gr.sendSmsg(watchMsg)
								}

							case protocol.UPC_Throw:
								{
									gr.Player[event.PlayerId].Status = PS_Lost
									gr.sendSmsg(&ServerMsg{
										Status:   PS_Lost,
										PlayerId: event.PlayerId,
									})
								}

							case protocol.UPC_Exit:
								{
									gr.Player[event.PlayerId].Status = PS_Exit
									gr.sendSmsg(&ServerMsg{
										Status:   PS_Exit,
										PlayerId: event.PlayerId,
									})
								}
							default:
								{
									gr.sendSmsg(&ServerMsg{
										Status: PS_Warning,
										SendId: event.PlayerId,
										ErrStr: protocol.NewErr(protocol.ErrPlayerCmdRejectCode).Error(),
									})
								}

							}
						}
					}
				case <-gr.Timer.C:
					{
						//超时就弃牌
						logger.Debug("timeout")
						curUser.Status = PS_Lost
						return nil
					}

				}
				return nil
			}

		}(); err != nil {
			logger.Error(err)
		}

		//没有输或者退出 就放回牌桌
		if curUser.Status < PS_Lost {
			gr.TablePlayer.PushBack(curUserId)
		}

		gr.checkExit()
	}

}

func addUser(uid int64, conn *websocket.Conn) {
	user := UserContext{
		Conn: conn,
		Uid:uid,
		RecvChanMap: map[protocol.EventType]chan []byte{},
		SendChanMap: map[protocol.EventType]chan []byte{},
	}
	ctx, cancel := context.WithCancel(context.Background())

	user.Cancel=cancel
	user.RecvChanMap[protocol.UC_SignCmd]=make(chan []byte,1024)
	user.RecvChanMap[protocol.UC_RoomCmd]=make(chan []byte,1024)
	user.RecvChanMap[protocol.UC_PlayCmd]=make(chan []byte,1024)
	user.RecvChanMap[protocol.UC_NotifyCmd]=make(chan []byte,1024)

	user.SendChanMap[protocol.UC_ErrCmd]=make(chan []byte,1024)
	user.SendChanMap[protocol.UC_SignCmd]=make(chan []byte,1024)
	user.SendChanMap[protocol.UC_RoomCmd]=make(chan []byte,1024)
	user.SendChanMap[protocol.UC_PlayCmd]=make(chan []byte,1024)
	user.SendChanMap[protocol.UC_NotifyCmd]=make(chan []byte,1024)

	go user.ProcessMsg(ctx)
	go user.RecvMsg(ctx)
	go user.SendMsg(ctx)

	UserMap.Store(uid, user)

	timer:=time.NewTimer(time.Second*10)
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case t:=<-timer.C:
			{
				logger.Debug(t)
			}
		}
	}
}
func delUser(uid int64) {
	if uptr,exist:=UserMap.Load(uid);exist{
		if user,ok:=uptr.(*UserContext);ok{
			user.Cancel()
			UserMap.Delete(uid)
		}
	}
}
func getUser(uid int64) (*UserContext, bool) {
	v, exist := UserMap.Load(uid)
	if !exist {
		return nil, false
	}
	c, ok := v.(*UserContext)
	if !ok {
		return nil, false
	}
	return c, true
}

func login(conn *websocket.Conn) (uid int64, err error) {
	lreq := protocol.WsRequest{}

	if err = conn.ReadJSON(&lreq); err != nil {
		logger.Warnning(err)
		return
	}
	//检查
	user := model.User{
		Email: "",
	}
	db := orm.Get()

	if err = db.Model(&model.User{}).Where(&user).Last(&user).Error; err != nil {
		//未找到用户
		return 0, protocol.NewErr(protocol.ErrNotFoundUserCode)
	}
	if user.Password != util.EncryptPassword("") {
		//密码不正确
		return 0, protocol.NewErr(protocol.ErrWrongPasswordCode)
	}

	return user.Id, nil
}

type UserStatus int

const (
	Offline UserStatus = iota
	Online
	RoomWait
	RoomPlay
)

func Cheat3(conn *websocket.Conn) {

	uid, err := login(conn)
	if err != nil {
		logger.Warnning(err)
		return
	}
	addUser(uid, conn)
	defer delUser(uid)
}

func play(room GameRoom) {

}
