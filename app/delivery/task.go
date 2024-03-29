package delivery

import (
	"net/http"
	"strconv"

	"Kanbanboard/app/delivery/middleware"
	"Kanbanboard/app/helper"
	"Kanbanboard/domain"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type TaskHandler struct {
	taskUsecase domain.TaskUsecase
}

func NewTaskHandler(r *gin.RouterGroup, taskUsecase domain.TaskUsecase) {
	handler := &TaskHandler{taskUsecase}
	taskRoute := r.Group("/tasks")
	taskRoute.Use(middleware.Authentication())
	taskRoute.GET("/", handler.GetTasks)
	taskRoute.Use(middleware.Authorization([]string{"user"}))
	taskRoute.POST("/", handler.StoreTask)
	taskRoute.PUT("/:taskId", handler.UpdateTask)
	taskRoute.PATCH("/update-status/:taskId", handler.UpdateStatusTask)
	taskRoute.PATCH("/update-category/:taskId", handler.UpdateCategoryTask)
	taskRoute.DELETE(":taskId", handler.DeleteTask)
}

func (t *TaskHandler) GetTasks(ctx *gin.Context) {

	tasks, err := t.taskUsecase.GetTasks(ctx.Request.Context())
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK,
		"data": tasks,
	})

}

func (t *TaskHandler) StoreTask(ctx *gin.Context) {
	type StoreTask struct {
		Title       string `json:"title" validate:"required"`
		Description string `json:"description" validate:"required"`
		CategoryID  int64  `json:"category_id" validate:"required"`
	}
	var storeTask StoreTask
	err := ctx.ShouldBindJSON(&storeTask)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	err = helper.ValidateStruct(storeTask)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	var task domain.Task
	userAuth := ctx.MustGet("user").(jwt.MapClaims)
	userID := int64(userAuth["id"].(float64))
	task.UserID = userID
	task.Status = false
	copier.Copy(&task, &storeTask)
	taskData, err := t.taskUsecase.StoreTask(ctx.Request.Context(), &task)
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"code": http.StatusCreated,
		"data": gin.H{
			"id":          taskData.ID,
			"title":       taskData.Title,
			"status":      taskData.Status,
			"description": taskData.Description,
			"user_id":     taskData.UserID,
			"category_id": taskData.CategoryID,
			"created_at":  taskData.CreatedAt,
		},
	})

}

func (t *TaskHandler) UpdateTask(ctx *gin.Context) {
	type UpdateTask struct {
		Title       string `json:"title" validate:"required"`
		Description string `json:"description" validate:"required"`
	}
	var updateTask UpdateTask
	err := ctx.ShouldBindJSON(&updateTask)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	err = helper.ValidateStruct(updateTask)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var task domain.Task

	taskId, _ := strconv.ParseInt(ctx.Param("taskId"), 10, 64)
	task.ID = taskId
	task, err = t.taskUsecase.GetTaskByID(ctx.Request.Context(), taskId)

	userAuth := ctx.MustGet("user").(jwt.MapClaims)
	userID := int64(userAuth["id"].(float64))

	if task.UserID != userID {
		ctx.JSON(getStatusCode(err), gin.H{"message": domain.ErrUnauthorized.Error()})
		return
	}

	copier.Copy(&task, &updateTask)
	taskData, err := t.taskUsecase.UpdateTask(ctx, &task)
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"id":          taskData.ID,
			"title":       taskData.Title,
			"description": taskData.Description,
			"status":      taskData.Status,
			"user_id":     taskData.UserID,
			"category_id": taskData.CategoryID,
			"updated_at":  taskData.UpdatedAt,
		},
	})

}

func (t *TaskHandler) UpdateStatusTask(ctx *gin.Context) {
	type UpdateStatusTask struct {
		Status bool `json:"status" validate:"required"`
	}

	var updateStatusTask UpdateStatusTask
	err := ctx.ShouldBindJSON(&updateStatusTask)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = helper.ValidateStruct(updateStatusTask)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var task domain.Task

	taskId, _ := strconv.ParseInt(ctx.Param("taskId"), 10, 64)
	task.ID = taskId
	task, err = t.taskUsecase.GetTaskByID(ctx.Request.Context(), taskId)

	copier.Copy(&task, &updateStatusTask)

	taskData, err := t.taskUsecase.UpdateTask(ctx, &task)
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"id":          taskData.ID,
			"title":       taskData.Title,
			"description": taskData.Description,
			"status":      taskData.Status,
			"user_id":     taskData.UserID,
			"category_id": taskData.CategoryID,
			"updated_at":  taskData.UpdatedAt,
		},
	})

}

func (t *TaskHandler) UpdateCategoryTask(ctx *gin.Context) {
	type UpdateCategoryTask struct {
		CategoryID int64 `json:"category_id" validate:"required"`
	}

	var updateCategoryTask UpdateCategoryTask
	err := ctx.ShouldBindJSON(&updateCategoryTask)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	err = helper.ValidateStruct(updateCategoryTask)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	var task domain.Task

	taskId, _ := strconv.ParseInt(ctx.Param("taskId"), 10, 64)
	task.ID = taskId
	task, err = t.taskUsecase.GetTaskByID(ctx.Request.Context(), taskId)

	copier.Copy(&task, &updateCategoryTask)

	taskData, err := t.taskUsecase.UpdateTask(ctx, &task)
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"id":          taskData.ID,
			"title":       taskData.Title,
			"description": taskData.Description,
			"status":      taskData.Status,
			"user_id":     taskData.UserID,
			"category_id": taskData.CategoryID,
			"updated_at":  taskData.UpdatedAt,
		},
	})

}

func (t *TaskHandler) DeleteTask(ctx *gin.Context) {
	taskId, _ := strconv.ParseInt(ctx.Param("taskId"), 10, 64)
	err := t.taskUsecase.DeleteTask(ctx, taskId)
	if err != nil {
		ctx.JSON(getStatusCode(err), gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "task has been successfully deleted"})

}
