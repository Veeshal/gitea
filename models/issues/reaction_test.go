// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package issues

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/setting"

	"github.com/stretchr/testify/assert"
)

func addReaction(t *testing.T, doerID, issueID, commentID int64, content string) {
	var reaction *Reaction
	var err error
	if commentID == 0 {
		reaction, err = CreateIssueReaction(doerID, issueID, content)
	} else {
		reaction, err = CreateCommentReaction(doerID, issueID, commentID, content)
	}
	assert.NoError(t, err)
	assert.NotNil(t, reaction)
}

func TestIssueAddReaction(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	var issue1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, 0, "heart")

	unittest.AssertExistsAndLoadBean(t, &Reaction{Type: "heart", UserID: user1.ID, IssueID: issue1ID})
}

func TestIssueAddDuplicateReaction(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	var issue1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, 0, "heart")

	reaction, err := CreateReaction(&ReactionOptions{
		DoerID:  user1.ID,
		IssueID: issue1ID,
		Type:    "heart",
	})
	assert.Error(t, err)
	assert.Equal(t, ErrReactionAlreadyExist{Reaction: "heart"}, err)

	existingR := unittest.AssertExistsAndLoadBean(t, &Reaction{Type: "heart", UserID: user1.ID, IssueID: issue1ID}).(*Reaction)
	assert.Equal(t, existingR.ID, reaction.ID)
}

func TestIssueDeleteReaction(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	var issue1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, 0, "heart")

	err := DeleteIssueReaction(user1.ID, issue1ID, "heart")
	assert.NoError(t, err)

	unittest.AssertNotExistsBean(t, &Reaction{Type: "heart", UserID: user1.ID, IssueID: issue1ID})
}

func TestIssueReactionCount(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	setting.UI.ReactionMaxUserNum = 2

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2}).(*user_model.User)
	user3 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3}).(*user_model.User)
	user4 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4}).(*user_model.User)
	ghost := user_model.NewGhostUser()

	var issueID int64 = 2
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1}).(*repo_model.Repository)

	addReaction(t, user1.ID, issueID, 0, "heart")
	addReaction(t, user2.ID, issueID, 0, "heart")
	addReaction(t, user3.ID, issueID, 0, "heart")
	addReaction(t, user3.ID, issueID, 0, "+1")
	addReaction(t, user4.ID, issueID, 0, "+1")
	addReaction(t, user4.ID, issueID, 0, "heart")
	addReaction(t, ghost.ID, issueID, 0, "-1")

	reactionsList, _, err := FindReactions(db.DefaultContext, FindReactionsOptions{
		IssueID: issueID,
	})
	assert.NoError(t, err)
	assert.Len(t, reactionsList, 7)
	_, err = reactionsList.LoadUsers(db.DefaultContext, repo)
	assert.NoError(t, err)

	reactions := reactionsList.GroupByType()
	assert.Len(t, reactions["heart"], 4)
	assert.Equal(t, 2, reactions["heart"].GetMoreUserCount())
	assert.Equal(t, user1.DisplayName()+", "+user2.DisplayName(), reactions["heart"].GetFirstUsers())
	assert.True(t, reactions["heart"].HasUser(1))
	assert.False(t, reactions["heart"].HasUser(5))
	assert.False(t, reactions["heart"].HasUser(0))
	assert.Len(t, reactions["+1"], 2)
	assert.Equal(t, 0, reactions["+1"].GetMoreUserCount())
	assert.Len(t, reactions["-1"], 1)
}

func TestIssueCommentAddReaction(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	var issue1ID int64 = 1
	var comment1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, comment1ID, "heart")

	unittest.AssertExistsAndLoadBean(t, &Reaction{Type: "heart", UserID: user1.ID, IssueID: issue1ID, CommentID: comment1ID})
}

func TestIssueCommentDeleteReaction(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2}).(*user_model.User)
	user3 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3}).(*user_model.User)
	user4 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4}).(*user_model.User)

	var issue1ID int64 = 1
	var comment1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, comment1ID, "heart")
	addReaction(t, user2.ID, issue1ID, comment1ID, "heart")
	addReaction(t, user3.ID, issue1ID, comment1ID, "heart")
	addReaction(t, user4.ID, issue1ID, comment1ID, "+1")

	reactionsList, _, err := FindReactions(db.DefaultContext, FindReactionsOptions{
		IssueID:   issue1ID,
		CommentID: comment1ID,
	})
	assert.NoError(t, err)
	assert.Len(t, reactionsList, 4)

	reactions := reactionsList.GroupByType()
	assert.Len(t, reactions["heart"], 3)
	assert.Len(t, reactions["+1"], 1)
}

func TestIssueCommentReactionCount(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1}).(*user_model.User)

	var issue1ID int64 = 1
	var comment1ID int64 = 1

	addReaction(t, user1.ID, issue1ID, comment1ID, "heart")
	assert.NoError(t, DeleteCommentReaction(user1.ID, issue1ID, comment1ID, "heart"))

	unittest.AssertNotExistsBean(t, &Reaction{Type: "heart", UserID: user1.ID, IssueID: issue1ID, CommentID: comment1ID})
}
