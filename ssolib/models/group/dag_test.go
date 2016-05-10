package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddOwnAsMemberWillFail(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)
	ga, err := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	assert.Nil(t, err)
	err = ga.AddGroupMember(th.Ctx, ga, ADMIN)
	assert.Equal(t, ErrGroupIncludingFailed, err)
}

/*
A->B->C->D
A->C
B->D
*/
func TestAddGroupMembersWorksAndIncreasingDepthAutomaticly(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)
	ga, err := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	assert.Nil(t, err)
	depthA, _ := getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 1, depthA)
	gb, err := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	depthB, _ := getGroupDepth(th.Ctx, gb.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, depthB)
	t.Log(th.Ctx.Lock)
	err = ga.AddGroupMember(th.Ctx, gb, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 2, depthA)
	gc, _ := CreateGroup(th.Ctx, &Group{
		Name: "C",
	})
	gd, _ := CreateGroup(th.Ctx, &Group{
		Name: "D",
	})
	gc.AddGroupMember(th.Ctx, gd, ADMIN)
	depthC, _ := getGroupDepth(th.Ctx, gc.Id)
	depthD, _ := getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)
	err = ga.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gd, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 4, depthA)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)
}

func TestAddGroupWillFailWhenCycleOccurs(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)
	ga, _ := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	gb, _ := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	ga.AddGroupMember(th.Ctx, gb, ADMIN)
	gc, _ := CreateGroup(th.Ctx, &Group{
		Name: "C",
	})
	gd, _ := CreateGroup(th.Ctx, &Group{
		Name: "D",
	})
	gc.AddGroupMember(th.Ctx, gd, ADMIN)
	ga.AddGroupMember(th.Ctx, gc, ADMIN)
	gb.AddGroupMember(th.Ctx, gd, ADMIN)
	gb.AddGroupMember(th.Ctx, gc, ADMIN)

	err := gd.AddGroupMember(th.Ctx, ga, ADMIN)
	assert.Equal(t, ErrGroupIncludingFailed, err)
	err = gd.AddGroupMember(th.Ctx, gb, ADMIN)
	assert.Equal(t, ErrGroupIncludingFailed, err)
	err = gd.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Equal(t, ErrGroupIncludingFailed, err)

}

func TestAddGroupWillFailWhenExceedsMaxDepth(t *testing.T) {
	EnableNestedGroup()
	SetMaxDepth(3)
	defer SetMaxDepth(0)
	th := NewTestHelper(t)
	ga, err := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	assert.Nil(t, err)
	depthA, _ := getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 1, depthA)
	gb, err := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	depthB, _ := getGroupDepth(th.Ctx, gb.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, depthB)
	t.Log(th.Ctx.Lock)
	err = ga.AddGroupMember(th.Ctx, gb, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 2, depthA)
	gc, _ := CreateGroup(th.Ctx, &Group{
		Name: "C",
	})
	gd, _ := CreateGroup(th.Ctx, &Group{
		Name: "D",
	})
	gc.AddGroupMember(th.Ctx, gd, ADMIN)
	depthC, _ := getGroupDepth(th.Ctx, gc.Id)
	depthD, _ := getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)
	err = ga.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gd, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Equal(t, err, ErrGroupIncludingFailed)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

}

func TestRemoveGroupMembersWorksAndDecreasingDepthAutomaticly(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)
	ga, _ := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	gb, _ := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	ga.AddGroupMember(th.Ctx, gb, ADMIN)
	gc, _ := CreateGroup(th.Ctx, &Group{
		Name: "C",
	})
	gd, _ := CreateGroup(th.Ctx, &Group{
		Name: "D",
	})
	gc.AddGroupMember(th.Ctx, gd, ADMIN)
	ga.AddGroupMember(th.Ctx, gc, ADMIN)
	gb.AddGroupMember(th.Ctx, gd, ADMIN)
	gb.AddGroupMember(th.Ctx, gc, ADMIN)

	err := gb.RemoveGroupMember(th.Ctx, gc)
	assert.Nil(t, err)
	depthA, _ := getGroupDepth(th.Ctx, ga.Id)
	depthB, _ := getGroupDepth(th.Ctx, gb.Id)
	depthC, _ := getGroupDepth(th.Ctx, gc.Id)
	depthD, _ := getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 3, depthA)
	assert.Equal(t, 2, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = gb.RemoveGroupMember(th.Ctx, gd)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 3, depthA)
	assert.Equal(t, 1, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = ga.RemoveGroupMember(th.Ctx, gc)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthA)
	assert.Equal(t, 1, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = gb.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 4, depthA)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = ga.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 4, depthA)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = gb.AddGroupMember(th.Ctx, gd, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 4, depthA)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = gc.RemoveGroupMember(th.Ctx, gd)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 3, depthA)
	assert.Equal(t, 2, depthB)
	assert.Equal(t, 1, depthC)
	assert.Equal(t, 1, depthD)

	err = gc.AddGroupMember(th.Ctx, gd, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 4, depthA)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

}

func TestDeleteGroupWorks(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)
	ga, err := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	assert.Nil(t, err)
	depthA, _ := getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 1, depthA)
	gb, err := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	depthB, _ := getGroupDepth(th.Ctx, gb.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, depthB)
	t.Log(th.Ctx.Lock)
	err = ga.AddGroupMember(th.Ctx, gb, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 2, depthA)
	gc, _ := CreateGroup(th.Ctx, &Group{
		Name: "C",
	})
	gd, _ := CreateGroup(th.Ctx, &Group{
		Name: "D",
	})
	gc.AddGroupMember(th.Ctx, gd, ADMIN)
	depthC, _ := getGroupDepth(th.Ctx, gc.Id)
	depthD, _ := getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)
	err = ga.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gd, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	err = gb.AddGroupMember(th.Ctx, gc, ADMIN)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 4, depthA)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthC, _ = getGroupDepth(th.Ctx, gc.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 3, depthB)
	assert.Equal(t, 2, depthC)
	assert.Equal(t, 1, depthD)

	err = DeleteGroup(th.Ctx, gc)
	assert.Nil(t, err)
	depthA, _ = getGroupDepth(th.Ctx, ga.Id)
	assert.Equal(t, 3, depthA)
	depthB, _ = getGroupDepth(th.Ctx, gb.Id)
	depthD, _ = getGroupDepth(th.Ctx, gd.Id)
	assert.Equal(t, 2, depthB)
	assert.Equal(t, 1, depthD)

}
