package service

import (
	"cheat/logger"
	"cheat/model"
	"cheat/model/plate"
	"cheat/orm"
	"cheat/protocol"
	"cheat/util"
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

var UserMap sync.Map

type UserContext struct {
	Conn *websocket.Conn

	RecvChanMap map[protocol.EventType]chan []byte
	SendChanMap map[protocol.EventType]chan []byte

	CmdChan chan int64
	RoomId  int64
}

func (u *UserContext) JoinRoom(rid int64) {
	u.RoomId = rid
}
func (u *UserContext) SendPlayMsg(data []byte) {
	u.SendChanMap[protocol.UC_PlayCmd] <- data
}
func (u *UserContext) ExitRoom() {
	u.CmdChan <- u.RoomId
	u.RoomId = 0
}



type UserPlayStatus int

const (
	UPS_Error      = iota
	UPS_Ready      //准备
	UPS_Wait_Plate //等待发牌
	UPS_Current    //轮到
	UPS_Wait_Cheat //等待下注
	UPS_Win_All    //赢了
	UPS_Lost       //输了
	UPS_Exit       //离开
	UPS_Warning    //警告

	UPS_Close //房间关闭
)

type UserWatchStatus int

const (
	UWS_Open = iota
	UWS_Close
)

type RoomPlayer struct {
	Id     int64
	Amount uint64
	Status UserPlayStatus
	Watch  UserWatchStatus
	Plate  *plate.ThreePlate
}

type UserPlayCmd int

const (
	UPC_Error = iota
	UPC_Ready
	UPC_Watch
	UPC_Cheat
	UPC_Check
	UPC_Throw
	UPC_Exit
)

type PlayerEvent struct {
	PlayCmd UserPlayCmd

	//玩家
	PlayerId int64

	//投注额
	Amount uint64

	//是否已看牌
	Watch UserWatchStatus

	//投注价值
	Volume uint64

	//是否查人
	CheckId int64

	//是否有人输了
	LoserId int64
}

type ServerMsg struct {
	//通知给玩家的状态
	PlayStatus UserPlayStatus

	//这人的id
	PlayerId int64

	//事件
	Event PlayerEvent

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
	PlayerChan chan *PlayerEvent
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

func (gr *GameRoom) generalPlate(count int) []*plate.ThreePlate {

	return []*plate.ThreePlate{}
}
func (gr *GameRoom) sendSmsg(msg *ServerMsg) {
	gr.ServerChan <- msg
}

func (gr *GameRoom) recvCmsg() *PlayerEvent {
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
		event := PlayerEvent{}
		if err := json.Unmarshal(jstr, &event); err != nil {
			logger.Warnning(err)
			continue
		}
		//检查发送过来的消息是否正确
		event.PlayerId = uid
		event.LoserId = 0
		event.Watch = gr.Player[uid].Watch
		//额度是否在限制内
		//event.Amount
		if event.PlayCmd == UPC_Error {
			//未知的指令
			continue
		}
		//查人是否在房间内
		if event.CheckId != 0 {
			if _, exist := gr.Player[event.CheckId]; !exist {
				//该玩家不在房间内
				continue
			}
		}

		if event.PlayCmd == UPC_Exit {
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
func (gr *GameRoom) sendUids() {
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
		PlayStatus: UPS_Close,
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
		if p.Status == UPS_Exit {
			delete(gr.Player, p.Id)
		}
	}
}

func (gr *GameRoom) play() {
	chect_func := func(msg *PlayerEvent, u *RoomPlayer) error {
		//检查余额
		if msg.Amount > u.Amount {
			//没钱了
		}
		//检查金额

		//不看牌
		if u.Watch == UWS_Close {
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
		if u.Watch == UWS_Close {
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
	check_func := func(msg *PlayerEvent, u *RoomPlayer) error {

		if gr.Player[u.Id].Plate.Less(gr.Player[msg.CheckId].Plate) {
			return errors.New("less")
		}
		return nil
	}

	//设置所有人为 等待发牌,未看牌,
	for _, v := range gr.Player {
		v.Status = UPS_Ready
		v.Watch = UWS_Close
	}
	gr.sendSmsg(&ServerMsg{PlayStatus: UPS_Ready})
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

					case UPC_Cheat:
						{
							if err := chect_func(event, gr.Player[event.PlayerId]); err != nil {
								return err
							}
							gr.Player[event.PlayerId].Status = UPS_Wait_Plate
							omsg.Event.Amount = event.Amount
							omsg.Event.Volume = gr.CurVol
						}

					case UPC_Exit:
						{
							gr.Player[event.PlayerId].Status = UPS_Exit
						}
					default:
						{
							return protocol.NewErr(protocol.ErrPlayerCmdRejectCode)
						}
					}
					return nil
				}(); err != nil {
					omsg.PlayStatus = UPS_Warning
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
		if p.Status == UPS_Wait_Plate {
			gr.TablePlayer.PushBack(p.Id)
		}
	}

	plates := gr.generalPlate(gr.TablePlayer.Len())

	//发牌
	for i := 0; i < gr.TablePlayer.Len(); i++ {
		tmp := gr.TablePlayer.PopFront().(int64)
		gr.Player[tmp].Plate = plates[i]
	}

	for {

		curUserId := gr.TablePlayer.PopFront().(int64)
		curUser := gr.Player[curUserId]

		//已经输了
		if curUser.Status == UPS_Lost {
			continue
		}

		if gr.TablePlayer.Len() == 0 {
			//赢了

			gr.sendSmsg(&ServerMsg{
				PlayStatus: UPS_Win_All,
				PlayerId:   curUserId,
			})
			break
		}
		gr.Timer = time.NewTimer(TimeOut)

		//通知所有人 当前用户在干嘛
		gr.sendSmsg(&ServerMsg{
			PlayStatus: UPS_Current,
			PlayerId:   curUserId,
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
							omsg.PlayStatus = UPS_Wait_Cheat
							NeedNext := errors.New("needNext")
							if err := func() error {

								switch event.PlayCmd {
								case UPC_Cheat:
									{
										//押注

										if err := chect_func(event, curUser); err != nil {
											return err
										}
										omsg.Event.Amount = event.Amount
										omsg.Event.Volume = gr.CurVol
										omsg.Event.Watch = curUser.Watch

									}
								case UPC_Check:
									{
										//查人
										checkUser := gr.Player[event.CheckId]

										switch checkUser.Status {
										case UPS_Lost:
											{
												//查牌的人已经挂了
												return protocol.NewErr(protocol.ErrPlayerLosedCode)
											}
										case UPS_Wait_Plate:
											{
												//正常

												//查牌也要给钱的
												if err := chect_func(event, curUser); err != nil {
													return err
												}
												if err := check_func(event, curUser); err != nil {
													//比别人小
													gr.Player[curUserId].Status = UPS_Lost
													omsg.Event.LoserId = curUserId
												} else {
													gr.Player[event.CheckId].Status = UPS_Lost
													omsg.Event.LoserId = event.CheckId
												}
												omsg.Event.CheckId = event.CheckId
												omsg.Event.Amount = event.Amount
												omsg.Event.Volume = gr.CurVol
												omsg.Event.Watch = curUser.Watch

											}
										case UPS_Exit:
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
								case UPC_Watch:
									{
										//看牌
										watchMsg := &ServerMsg{
											PlayStatus: UPS_Wait_Cheat,
											Event: PlayerEvent{
												PlayCmd: UPC_Watch,
											},
											Plate:    curUser.Plate,
											PlayerId: curUser.Id,
											SendId:   curUser.Id,
										}
										gr.sendSmsg(watchMsg)
										//已经看牌
										gr.Player[curUserId].Watch = UWS_Open

										//结束
										return NeedNext
									}
								case UPC_Throw:
									{
										//弃牌
										gr.Player[curUserId].Status = UPS_Lost
										omsg.PlayStatus = UPS_Lost
									}
								case UPC_Exit:
									{
										//退出
										gr.Player[curUserId].Status = UPS_Exit
										omsg.PlayStatus = UPS_Exit
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
								omsg.PlayStatus = UPS_Warning
								omsg.ErrStr = err.Error()
								gr.sendSmsg(omsg)
							} else {
								//发送 给所有
								gr.sendSmsg(omsg)
							}
						} else {
							//不是当前人只能看和弃牌和退出
							switch event.PlayCmd {
							case UPC_Watch:
								{
									//看牌
									tuser := gr.Player[event.PlayerId]
									watchMsg := &ServerMsg{
										PlayStatus: UPS_Wait_Cheat,
										Event: PlayerEvent{
											PlayCmd: UPC_Watch,
										},
										Plate:    tuser.Plate,
										PlayerId: tuser.Id,
										SendId:   tuser.Id,
									}
									tuser.Watch = UWS_Open
									gr.sendSmsg(watchMsg)
								}

							case UPC_Throw:
								{
									gr.Player[event.PlayerId].Status = UPS_Lost
									gr.sendSmsg(&ServerMsg{
										PlayStatus: UPS_Lost,
										PlayerId:   event.PlayerId,
									})
								}

							case UPC_Exit:
								{
									gr.Player[event.PlayerId].Status = UPS_Exit
									gr.sendSmsg(&ServerMsg{
										PlayStatus: UPS_Exit,
										PlayerId:   event.PlayerId,
									})
								}
							default:
								{
									gr.sendSmsg(&ServerMsg{
										PlayStatus: UPS_Warning,
										SendId:     event.PlayerId,
										ErrStr:     protocol.NewErr(protocol.ErrPlayerCmdRejectCode).Error(),
									})
								}

							}
						}
					}
				case <-gr.Timer.C:
					{
						//超时就弃牌
						logger.Debug("timeout")
						curUser.Status = UPS_Lost
						return nil
					}

				}
				return nil
			}

		}(); err != nil {
			logger.Error(err)
		}

		//没有输或者退出 就放回牌桌
		if curUser.Status < UPS_Lost {
			gr.TablePlayer.PushBack(curUserId)
		}

		gr.checkExit()
	}

}

var RoomCount = uint64(0)

func joinUser(uid int64, conn *websocket.Conn) {
	user := UserContext{
		Conn: conn,
	}
	go func() {
		for {
			req := &protocol.WsRequest{}
			if err := conn.ReadJSON(req); err != nil {
				logger.Warnning(err)
			}
			switch req.Cmd {
			case protocol.UC_ErrCmd:
				continue
			case protocol.UC_SignCmd:
				user.RecvChanMap[protocol.UC_SignCmd] <- req.Data
			case protocol.UC_RoomCmd:
				user.RecvChanMap[protocol.UC_RoomCmd] <- req.Data
			case protocol.UC_PlayCmd:
				user.RecvChanMap[protocol.UC_PlayCmd] <- req.Data
			}
		}
	}()

	go func() {
		for {
			res := &protocol.WsResponse{}
			select {
			case data := <-user.SendChanMap[protocol.UC_SignCmd]:
				{
					res.Cmd = protocol.UC_SignCmd
					res.Data = data
				}
			case data := <-user.SendChanMap[protocol.UC_RoomCmd]:
				{
					res.Cmd = protocol.UC_RoomCmd
					res.Data = data
				}
			case data := <-user.SendChanMap[protocol.UC_PlayCmd]:
				{
					res.Cmd = protocol.UC_PlayCmd
					res.Data = data
				}
			}
		}
	}()

	UserMap.Store(uid, user)
}
func exitUser(uid int64) {
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
	joinUser(uid, conn)
	defer exitUser(uid)

	//main logic
}

func play(room GameRoom) {

}
